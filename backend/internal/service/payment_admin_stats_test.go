//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestAdminListOrdersFiltersByCreatedAtRange(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createPaymentStatsUser(t, ctx, client, "range@example.com")
	svc := &PaymentService{entClient: client}

	start := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)
	createPaymentStatsOrder(t, ctx, client, paymentStatsOrderInput{
		UserID: user.ID, UserEmail: user.Email, UserName: user.Username,
		OutTradeNo: "range-before", CreatedAt: start.Add(-time.Hour),
		Status: OrderStatusCompleted, Amount: 10, PayAmount: 10,
	})
	inRange := createPaymentStatsOrder(t, ctx, client, paymentStatsOrderInput{
		UserID: user.ID, UserEmail: user.Email, UserName: user.Username,
		OutTradeNo: "range-inside", CreatedAt: start.Add(6 * time.Hour),
		Status: OrderStatusCompleted, Amount: 20, PayAmount: 20,
	})
	createPaymentStatsOrder(t, ctx, client, paymentStatsOrderInput{
		UserID: user.ID, UserEmail: user.Email, UserName: user.Username,
		OutTradeNo: "range-after", CreatedAt: end.Add(time.Second),
		Status: OrderStatusCompleted, Amount: 30, PayAmount: 30,
	})

	orders, total, err := svc.AdminListOrders(ctx, 0, OrderListParams{
		Page: 1, PageSize: 20,
		StartTime: start, EndTime: end,
	})

	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, orders, 1)
	require.Equal(t, inRange.ID, orders[0].ID)
}

func TestGetAdminOrderStatsSummarizesProfitUsersAndSubscriptions(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	alice := createPaymentStatsUser(t, ctx, client, "alice@example.com")
	bob := createPaymentStatsUser(t, ctx, client, "bob@example.com")
	svc := &PaymentService{entClient: client}

	planBasic, err := client.SubscriptionPlan.Create().
		SetGroupID(10).
		SetName("Basic").
		SetDescription("Basic plan").
		SetPrice(30).
		SetValidityDays(30).
		SetValidityUnit("days").
		Save(ctx)
	require.NoError(t, err)
	planPro, err := client.SubscriptionPlan.Create().
		SetGroupID(20).
		SetName("Pro").
		SetDescription("Pro plan").
		SetPrice(80).
		SetValidityDays(30).
		SetValidityUnit("days").
		Save(ctx)
	require.NoError(t, err)

	start := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)
	createPaymentStatsOrder(t, ctx, client, paymentStatsOrderInput{
		UserID: alice.ID, UserEmail: alice.Email, UserName: alice.Username,
		OutTradeNo: "alice-balance", CreatedAt: start.Add(time.Hour),
		Status: OrderStatusCompleted, OrderType: payment.OrderTypeBalance,
		Amount: 100, PayAmount: 102, FeeRate: 2, RefundAmount: 10,
	})
	createPaymentStatsOrder(t, ctx, client, paymentStatsOrderInput{
		UserID: alice.ID, UserEmail: alice.Email, UserName: alice.Username,
		OutTradeNo: "alice-basic", CreatedAt: start.Add(2 * time.Hour),
		Status: OrderStatusCompleted, OrderType: payment.OrderTypeSubscription,
		PlanID: &planBasic.ID, Amount: 30, PayAmount: 33, FeeRate: 1,
	})
	createPaymentStatsOrder(t, ctx, client, paymentStatsOrderInput{
		UserID: bob.ID, UserEmail: bob.Email, UserName: bob.Username,
		OutTradeNo: "bob-pro", CreatedAt: start.Add(3 * time.Hour),
		Status: OrderStatusRefunded, OrderType: payment.OrderTypeSubscription,
		PlanID: &planPro.ID, Amount: 80, PayAmount: 88, FeeRate: 2.5, RefundAmount: 88,
	})
	createPaymentStatsOrder(t, ctx, client, paymentStatsOrderInput{
		UserID: bob.ID, UserEmail: bob.Email, UserName: bob.Username,
		OutTradeNo: "bob-pending", CreatedAt: start.Add(4 * time.Hour),
		Status: OrderStatusPending, OrderType: payment.OrderTypeBalance,
		Amount: 50, PayAmount: 50,
	})
	createPaymentStatsOrder(t, ctx, client, paymentStatsOrderInput{
		UserID: bob.ID, UserEmail: bob.Email, UserName: bob.Username,
		OutTradeNo: "bob-outside", CreatedAt: end.Add(time.Hour),
		Status: OrderStatusCompleted, OrderType: payment.OrderTypeBalance,
		Amount: 999, PayAmount: 999,
	})

	stats, err := svc.GetAdminOrderStats(ctx, OrderListParams{StartTime: start, EndTime: end})

	require.NoError(t, err)
	require.Equal(t, 4, stats.TotalOrders)
	require.Equal(t, 3, stats.PaidOrders)
	require.Equal(t, 1, stats.PendingOrders)
	require.InDelta(t, 210, stats.GrossAmount, 0.001)
	require.InDelta(t, 223, stats.GrossPayAmount, 0.001)
	require.InDelta(t, 88, stats.RefundAmount, 0.001)
	require.InDelta(t, 4.57, stats.FeeAmount, 0.001)
	require.InDelta(t, 130.43, stats.NetPayAmount, 0.001)

	require.Len(t, stats.TopUsers, 2)
	require.Equal(t, alice.ID, stats.TopUsers[0].UserID)
	require.Equal(t, "alice@example.com", stats.TopUsers[0].Email)
	require.InDelta(t, 135, stats.TopUsers[0].GrossPayAmount, 0.001)
	require.InDelta(t, 132.63, stats.TopUsers[0].NetPayAmount, 0.001)

	require.Len(t, stats.SubscriptionPlans, 2)
	require.Equal(t, planPro.ID, stats.SubscriptionPlans[0].PlanID)
	require.Equal(t, "Pro", stats.SubscriptionPlans[0].PlanName)
	require.InDelta(t, -2.2, stats.SubscriptionPlans[0].NetPayAmount, 0.001)
	require.Equal(t, planBasic.ID, stats.SubscriptionPlans[1].PlanID)
	require.Equal(t, "Basic", stats.SubscriptionPlans[1].PlanName)
	require.InDelta(t, 32.67, stats.SubscriptionPlans[1].NetPayAmount, 0.001)
}

func TestGetAdminOrderStatsOnlyDeductsFinalizedRefunds(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createPaymentStatsUser(t, ctx, client, "refund-states@example.com")
	svc := &PaymentService{entClient: client}

	start := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		tradeNo string
		status  string
		pay     float64
		refund  float64
	}{
		{tradeNo: "refund-pending", status: OrderStatusRefundPending, pay: 100, refund: 40},
		{tradeNo: "refund-failed", status: OrderStatusRefundFailed, pay: 80, refund: 80},
		{tradeNo: "refund-final", status: OrderStatusRefunded, pay: 60, refund: 60},
		{tradeNo: "refund-partial", status: OrderStatusPartiallyRefunded, pay: 50, refund: 10},
	} {
		createPaymentStatsOrder(t, ctx, client, paymentStatsOrderInput{
			UserID: user.ID, UserEmail: user.Email, UserName: user.Username,
			OutTradeNo: tc.tradeNo, CreatedAt: start.Add(time.Hour),
			Status: tc.status, Amount: tc.pay, PayAmount: tc.pay, RefundAmount: tc.refund,
		})
	}

	stats, err := svc.GetAdminOrderStats(ctx, OrderListParams{StartTime: start, EndTime: end})

	require.NoError(t, err)
	require.Equal(t, 4, stats.PaidOrders)
	require.InDelta(t, 290, stats.GrossPayAmount, 0.001)
	require.InDelta(t, 70, stats.RefundAmount, 0.001)
	require.InDelta(t, 220, stats.NetPayAmount, 0.001)
}

type paymentStatsOrderInput struct {
	UserID       int64
	UserEmail    string
	UserName     string
	OutTradeNo   string
	CreatedAt    time.Time
	Status       string
	OrderType    string
	PlanID       *int64
	Amount       float64
	PayAmount    float64
	FeeRate      float64
	RefundAmount float64
}

func createPaymentStatsUser(t *testing.T, ctx context.Context, client *dbent.Client, email string) *dbent.User {
	t.Helper()
	user, err := client.User.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetUsername(email).
		Save(ctx)
	require.NoError(t, err)
	return user
}

func createPaymentStatsOrder(t *testing.T, ctx context.Context, client *dbent.Client, input paymentStatsOrderInput) *dbent.PaymentOrder {
	t.Helper()
	orderType := input.OrderType
	if orderType == "" {
		orderType = payment.OrderTypeBalance
	}
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	paidAt := createdAt.Add(time.Minute)
	builder := client.PaymentOrder.Create().
		SetUserID(input.UserID).
		SetUserEmail(input.UserEmail).
		SetUserName(input.UserName).
		SetAmount(input.Amount).
		SetPayAmount(input.PayAmount).
		SetFeeRate(input.FeeRate).
		SetRefundAmount(input.RefundAmount).
		SetRechargeCode(input.OutTradeNo + "-code").
		SetOutTradeNo(input.OutTradeNo).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo(input.OutTradeNo + "-trade").
		SetOrderType(orderType).
		SetStatus(input.Status).
		SetExpiresAt(createdAt.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetCreatedAt(createdAt).
		SetUpdatedAt(createdAt)
	if input.Status != OrderStatusPending {
		builder.SetPaidAt(paidAt)
	}
	if input.PlanID != nil {
		builder.SetPlanID(*input.PlanID)
	}
	order, err := builder.Save(ctx)
	require.NoError(t, err)
	return order
}
