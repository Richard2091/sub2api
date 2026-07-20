import { describe, expect, it, beforeEach, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AdminOrdersView from '../AdminOrdersView.vue'

const getOrders = vi.hoisted(() => vi.fn())
const getOrderStats = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, fallback?: string) => fallback || key,
    }),
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getOrders,
    getOrderStats,
    getOrder: vi.fn(),
    cancelOrder: vi.fn(),
    retryRecharge: vi.fn(),
    refundOrder: vi.fn(),
    queryRefund: vi.fn(),
  },
  default: {
    getOrders,
    getOrderStats,
    getOrder: vi.fn(),
    cancelOrder: vi.fn(),
    retryRecharge: vi.fn(),
    refundOrder: vi.fn(),
    queryRefund: vi.fn(),
  },
}))

function mountView() {
  return mount(AdminOrdersView, {
    global: {
      stubs: {
        AppLayout: { template: '<main><slot /></main>' },
        DateRangePicker: {
          name: 'DateRangePicker',
          props: ['startDate', 'endDate'],
          template: '<button data-testid="date-range-picker" />',
        },
        Select: true,
        Icon: true,
        OrderTable: { props: ['orders', 'loading'], template: '<section data-testid="orders-table"><slot name="actions" :row="orders[0] || {}" /></section>' },
        Pagination: true,
        BaseDialog: true,
        AdminRefundDialog: true,
        OrderStatusBadge: true,
      },
    },
  })
}

describe('AdminOrdersView analytics', () => {
  beforeEach(() => {
    getOrders.mockResolvedValue({ data: { items: [], total: 0 } })
    getOrderStats.mockResolvedValue({
      data: {
        total_orders: 4,
        paid_orders: 3,
        pending_orders: 1,
        gross_amount: 210,
        gross_pay_amount: 223,
        refund_amount: 98,
        fee_amount: 4.57,
        net_pay_amount: 120.43,
        avg_pay_amount: 74.33,
        top_users: [
          {
            user_id: 1,
            email: 'alice@example.com',
            name: 'Alice',
            order_count: 2,
            gross_amount: 130,
            gross_pay_amount: 135,
            refund_amount: 10,
            fee_amount: 2.37,
            net_pay_amount: 122.63,
          },
        ],
        subscription_plans: [
          {
            plan_id: 2,
            plan_name: 'Pro',
            order_count: 1,
            gross_amount: 80,
            gross_pay_amount: 88,
            refund_amount: 88,
            fee_amount: 2.2,
            net_pay_amount: -2.2,
          },
        ],
      },
    })
  })

  it('loads order stats and renders profit summary plus distribution sections', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(getOrderStats).toHaveBeenCalled()
    expect(wrapper.text()).toContain('$120.43')
    expect(wrapper.text()).toContain('alice@example.com')
    expect(wrapper.text()).toContain('Pro')
  })

  it('reloads orders and stats with the selected date range', async () => {
    const wrapper = mountView()
    await flushPromises()

    getOrders.mockClear()
    getOrderStats.mockClear()
    wrapper.findComponent({ name: 'DateRangePicker' }).vm.$emit('change', {
      startDate: '2026-07-01',
      endDate: '2026-07-15',
      preset: null,
    })
    await flushPromises()

    expect(getOrders).toHaveBeenCalledWith(expect.objectContaining({
      start_date: '2026-07-01',
      end_date: '2026-07-15',
    }))
    expect(getOrderStats).toHaveBeenCalledWith(expect.objectContaining({
      start_date: '2026-07-01',
      end_date: '2026-07-15',
    }))
  })
})
