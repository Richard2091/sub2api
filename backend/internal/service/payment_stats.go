package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionplan"
)

// --- Dashboard & Analytics ---

func (s *PaymentService) GetDashboardStats(ctx context.Context, days int) (*DashboardStats, error) {
	if days <= 0 {
		days = 30
	}
	now := time.Now()
	since := now.AddDate(0, 0, -days)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	paidStatuses := []string{OrderStatusCompleted, OrderStatusPaid, OrderStatusRecharging}

	orders, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.StatusIn(paidStatuses...),
			paymentorder.PaidAtGTE(since),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	st := &DashboardStats{}
	computeBasicStats(st, orders, todayStart)

	st.PendingOrders, err = s.entClient.PaymentOrder.Query().
		Where(paymentorder.StatusEQ(OrderStatusPending)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	st.DailySeries = buildDailySeries(orders, since, days)
	st.PaymentMethods = buildMethodDistribution(orders)
	st.TopUsers = buildTopUsers(orders)

	return st, nil
}

func (s *PaymentService) GetAdminOrderStats(ctx context.Context, p OrderListParams) (*AdminOrderStats, error) {
	q := s.entClient.PaymentOrder.Query()
	if p.UserID > 0 {
		q = q.Where(paymentorder.UserIDEQ(p.UserID))
	}
	if p.Status != "" {
		q = q.Where(paymentorder.StatusEQ(p.Status))
	}
	if p.OrderType != "" {
		q = q.Where(paymentorder.OrderTypeEQ(p.OrderType))
	}
	if p.PaymentType != "" {
		q = q.Where(paymentorder.PaymentTypeEQ(p.PaymentType))
	}
	if !p.StartTime.IsZero() {
		q = q.Where(paymentorder.CreatedAtGTE(p.StartTime))
	}
	if !p.EndTime.IsZero() {
		q = q.Where(paymentorder.CreatedAtLT(p.EndTime))
	}
	if p.Keyword != "" {
		q = q.Where(paymentorder.Or(
			paymentorder.OutTradeNoContainsFold(p.Keyword),
			paymentorder.UserEmailContainsFold(p.Keyword),
			paymentorder.UserNameContainsFold(p.Keyword),
		))
	}

	orders, err := q.All(ctx)
	if err != nil {
		return nil, err
	}

	stats := &AdminOrderStats{TotalOrders: len(orders)}
	userStats := make(map[int64]*AdminOrderUserStat)
	planStats := make(map[int64]*AdminOrderSubscriptionStat)
	planIDs := make(map[int64]struct{})

	for _, order := range orders {
		if order.Status == OrderStatusPending {
			stats.PendingOrders++
		}
		if !isAdminRevenueOrder(order) {
			continue
		}

		stats.PaidOrders++
		addOrderAmounts(&stats.GrossAmount, &stats.GrossPayAmount, &stats.RefundAmount, &stats.FeeAmount, &stats.NetPayAmount, order)

		us, ok := userStats[order.UserID]
		if !ok {
			us = &AdminOrderUserStat{UserID: order.UserID, Email: order.UserEmail, Name: order.UserName}
			userStats[order.UserID] = us
		}
		us.OrderCount++
		addOrderAmounts(&us.GrossAmount, &us.GrossPayAmount, &us.RefundAmount, &us.FeeAmount, &us.NetPayAmount, order)

		if order.OrderType == "subscription" && order.PlanID != nil {
			planID := *order.PlanID
			ps, ok := planStats[planID]
			if !ok {
				ps = &AdminOrderSubscriptionStat{PlanID: planID}
				planStats[planID] = ps
				planIDs[planID] = struct{}{}
			}
			ps.OrderCount++
			addOrderAmounts(&ps.GrossAmount, &ps.GrossPayAmount, &ps.RefundAmount, &ps.FeeAmount, &ps.NetPayAmount, order)
		}
	}

	if stats.PaidOrders > 0 {
		stats.AvgPayAmount = stats.GrossPayAmount / float64(stats.PaidOrders)
	}
	roundAdminOrderStats(stats)
	stats.TopUsers = sortedAdminOrderUserStats(userStats)
	stats.SubscriptionPlans = sortedAdminOrderSubscriptionStats(planStats, loadPlanNames(ctx, s, planIDs))
	return stats, nil
}

func isAdminRevenueOrder(order *dbent.PaymentOrder) bool {
	if order == nil || order.PaidAt == nil {
		return false
	}
	switch order.Status {
	case OrderStatusPending, OrderStatusCancelled, OrderStatusExpired, OrderStatusFailed:
		return false
	default:
		return true
	}
}

func addOrderAmounts(grossAmount, grossPayAmount, refundAmount, feeAmount, netPayAmount *float64, order *dbent.PaymentOrder) {
	fee := order.PayAmount * order.FeeRate / 100
	finalizedRefund := finalizedRefundAmount(order)
	*grossAmount += order.Amount
	*grossPayAmount += order.PayAmount
	*refundAmount += finalizedRefund
	*feeAmount += fee
	*netPayAmount += order.PayAmount - finalizedRefund - fee
}

func finalizedRefundAmount(order *dbent.PaymentOrder) float64 {
	if order == nil {
		return 0
	}
	switch order.Status {
	case OrderStatusRefunded, OrderStatusPartiallyRefunded:
		return order.RefundAmount
	default:
		return 0
	}
}

func roundAdminOrderStats(stats *AdminOrderStats) {
	stats.GrossAmount = roundMoney(stats.GrossAmount)
	stats.GrossPayAmount = roundMoney(stats.GrossPayAmount)
	stats.RefundAmount = roundMoney(stats.RefundAmount)
	stats.FeeAmount = roundMoney(stats.FeeAmount)
	stats.NetPayAmount = roundMoney(stats.NetPayAmount)
	stats.AvgPayAmount = roundMoney(stats.AvgPayAmount)
}

func sortedAdminOrderUserStats(items map[int64]*AdminOrderUserStat) []AdminOrderUserStat {
	out := make([]AdminOrderUserStat, 0, len(items))
	for _, item := range items {
		item.GrossAmount = roundMoney(item.GrossAmount)
		item.GrossPayAmount = roundMoney(item.GrossPayAmount)
		item.RefundAmount = roundMoney(item.RefundAmount)
		item.FeeAmount = roundMoney(item.FeeAmount)
		item.NetPayAmount = roundMoney(item.NetPayAmount)
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].GrossPayAmount == out[j].GrossPayAmount {
			return out[i].UserID < out[j].UserID
		}
		return out[i].GrossPayAmount > out[j].GrossPayAmount
	})
	return out
}

func sortedAdminOrderSubscriptionStats(items map[int64]*AdminOrderSubscriptionStat, planNames map[int64]string) []AdminOrderSubscriptionStat {
	out := make([]AdminOrderSubscriptionStat, 0, len(items))
	for _, item := range items {
		item.PlanName = planNames[item.PlanID]
		if item.PlanName == "" {
			item.PlanName = "Plan #" + strconv.FormatInt(item.PlanID, 10)
		}
		item.GrossAmount = roundMoney(item.GrossAmount)
		item.GrossPayAmount = roundMoney(item.GrossPayAmount)
		item.RefundAmount = roundMoney(item.RefundAmount)
		item.FeeAmount = roundMoney(item.FeeAmount)
		item.NetPayAmount = roundMoney(item.NetPayAmount)
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].GrossPayAmount == out[j].GrossPayAmount {
			return out[i].PlanID < out[j].PlanID
		}
		return out[i].GrossPayAmount > out[j].GrossPayAmount
	})
	return out
}

func loadPlanNames(ctx context.Context, s *PaymentService, ids map[int64]struct{}) map[int64]string {
	if len(ids) == 0 {
		return map[int64]string{}
	}
	planIDs := make([]int64, 0, len(ids))
	for id := range ids {
		planIDs = append(planIDs, id)
	}
	plans, err := s.entClient.SubscriptionPlan.Query().Where(subscriptionplan.IDIn(planIDs...)).All(ctx)
	if err != nil {
		return map[int64]string{}
	}
	names := make(map[int64]string, len(plans))
	for _, plan := range plans {
		names[plan.ID] = plan.Name
	}
	return names
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func computeBasicStats(st *DashboardStats, orders []*dbent.PaymentOrder, todayStart time.Time) {
	st.TotalAmount = make(CurrencyAmounts)
	st.TodayAmount = make(CurrencyAmounts)
	st.AvgAmount = make(CurrencyAmounts)
	currencyCounts := make(map[string]int)
	var todayCount int
	for _, o := range orders {
		currency := PaymentOrderCurrency(o)
		st.TotalAmount[currency] += o.PayAmount
		currencyCounts[currency]++
		if o.PaidAt != nil && !o.PaidAt.Before(todayStart) {
			st.TodayAmount[currency] += o.PayAmount
			todayCount++
		}
	}
	st.TotalCount = len(orders)
	st.TodayCount = todayCount
	for currency, totalAmount := range st.TotalAmount {
		st.AvgAmount[currency] = roundAmount(totalAmount / float64(currencyCounts[currency]))
	}
	roundCurrencyAmounts(st.TotalAmount)
	roundCurrencyAmounts(st.TodayAmount)
}

func buildDailySeries(orders []*dbent.PaymentOrder, since time.Time, days int) []DailyStats {
	dailyMap := make(map[string]*DailyStats)
	for _, o := range orders {
		if o.PaidAt == nil {
			continue
		}
		date := o.PaidAt.Format("2006-01-02")
		ds, ok := dailyMap[date]
		if !ok {
			ds = &DailyStats{Date: date, Amount: make(CurrencyAmounts)}
			dailyMap[date] = ds
		}
		ds.Amount[PaymentOrderCurrency(o)] += o.PayAmount
		ds.Count++
	}
	series := make([]DailyStats, 0, days)
	for i := 0; i < days; i++ {
		date := since.AddDate(0, 0, i+1).Format("2006-01-02")
		if ds, ok := dailyMap[date]; ok {
			roundCurrencyAmounts(ds.Amount)
			series = append(series, *ds)
		} else {
			series = append(series, DailyStats{Date: date, Amount: make(CurrencyAmounts)})
		}
	}
	return series
}

func buildMethodDistribution(orders []*dbent.PaymentOrder) []PaymentMethodStat {
	methodMap := make(map[string]*PaymentMethodStat)
	for _, o := range orders {
		ms, ok := methodMap[o.PaymentType]
		if !ok {
			ms = &PaymentMethodStat{Type: o.PaymentType, Amount: make(CurrencyAmounts)}
			methodMap[o.PaymentType] = ms
		}
		ms.Amount[PaymentOrderCurrency(o)] += o.PayAmount
		ms.Count++
	}
	methods := make([]PaymentMethodStat, 0, len(methodMap))
	for _, ms := range methodMap {
		roundCurrencyAmounts(ms.Amount)
		methods = append(methods, *ms)
	}
	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Type < methods[j].Type
	})
	return methods
}

func buildTopUsers(orders []*dbent.PaymentOrder) TopUsersByCurrency {
	userMap := make(map[string]map[int64]*TopUserStat)
	for _, o := range orders {
		currency := PaymentOrderCurrency(o)
		users, ok := userMap[currency]
		if !ok {
			users = make(map[int64]*TopUserStat)
			userMap[currency] = users
		}
		us, ok := users[o.UserID]
		if !ok {
			us = &TopUserStat{UserID: o.UserID, Email: o.UserEmail}
			users[o.UserID] = us
		}
		us.Amount += o.PayAmount
	}
	result := make(TopUsersByCurrency, len(userMap))
	for currency, users := range userMap {
		userList := make([]*TopUserStat, 0, len(users))
		for _, us := range users {
			us.Amount = roundAmount(us.Amount)
			userList = append(userList, us)
		}
		sort.Slice(userList, func(i, j int) bool {
			return userList[i].Amount > userList[j].Amount
		})
		limit := topUsersLimit
		if len(userList) < limit {
			limit = len(userList)
		}
		result[currency] = make([]TopUserStat, 0, limit)
		for i := 0; i < limit; i++ {
			result[currency] = append(result[currency], *userList[i])
		}
	}
	return result
}

func roundCurrencyAmounts(amounts CurrencyAmounts) {
	for currency, amount := range amounts {
		amounts[currency] = roundAmount(amount)
	}
}

func roundAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}

// --- Audit Logs ---

func (s *PaymentService) writeAuditLog(ctx context.Context, oid int64, action, op string, detail map[string]any) {
	dj, _ := json.Marshal(detail)
	_, err := s.entClient.PaymentAuditLog.Create().SetOrderID(strconv.FormatInt(oid, 10)).SetAction(action).SetDetail(string(dj)).SetOperator(op).Save(ctx)
	if err != nil {
		slog.Error("audit log failed", "orderID", oid, "action", action, "error", err)
	}
}

func (s *PaymentService) GetOrderAuditLogs(ctx context.Context, oid int64) ([]*dbent.PaymentAuditLog, error) {
	return s.entClient.PaymentAuditLog.Query().Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(oid, 10))).Order(paymentauditlog.ByCreatedAt()).All(ctx)
}
