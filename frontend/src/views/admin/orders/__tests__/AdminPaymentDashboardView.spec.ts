import { describe, expect, it, beforeEach, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AdminPaymentDashboardView from '../AdminPaymentDashboardView.vue'

const getDashboard = vi.hoisted(() => vi.fn())

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
  }),
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getDashboard,
  },
  default: {
    getDashboard,
  },
}))

function mountView() {
  return mount(AdminPaymentDashboardView, {
    global: {
      stubs: {
        AppLayout: { template: '<main><slot /></main>' },
        LoadingSpinner: true,
        DateRangePicker: {
          props: ['startDate', 'endDate'],
          emits: ['change', 'update:startDate', 'update:endDate'],
          template: `
            <button
              data-testid="date-range-picker"
              @click="$emit('change', { startDate: '2026-07-01', endDate: '2026-07-15', preset: null })"
            />
          `,
        },
        DailyRevenueChart: true,
        Icon: true,
      },
    },
  })
}

describe('AdminPaymentDashboardView profit analytics', () => {
  beforeEach(() => {
    getDashboard.mockResolvedValue({
      data: {
        today_amount: 105,
        total_amount: 147,
        today_count: 1,
        total_count: 2,
        avg_amount: 73.5,
        daily_series: [],
        payment_methods: [{ type: 'alipay', amount: 147, count: 2 }],
        top_users: [],
        profit_stats: {
          total_orders: 2,
          paid_orders: 2,
          pending_orders: 0,
          gross_amount: 140,
          gross_pay_amount: 147,
          refund_amount: 42,
          fee_amount: 2.1,
          net_pay_amount: 102.9,
          avg_pay_amount: 73.5,
          top_users: [
            {
              user_id: 1,
              email: 'alice@example.com',
              name: 'Alice',
              order_count: 2,
              gross_amount: 140,
              gross_pay_amount: 147,
              refund_amount: 42,
              fee_amount: 2.1,
              net_pay_amount: 102.9,
            },
          ],
          subscription_plans: [
            {
              plan_id: 2,
              plan_name: 'Pro',
              order_count: 1,
              gross_amount: 40,
              gross_pay_amount: 42,
              refund_amount: 42,
              fee_amount: 0,
              net_pay_amount: 0,
            },
          ],
        },
      },
    })
  })

  it('renders profit analytics from the payment dashboard response', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(getDashboard).toHaveBeenCalledWith({
      start_date: expect.any(String),
      end_date: expect.any(String),
    })
    expect(wrapper.text()).toContain('$102.90')
    expect(wrapper.text()).toContain('alice@example.com')
    expect(wrapper.text()).toContain('Pro')
  })

  it('reloads dashboard stats with the selected custom date range', async () => {
    const wrapper = mountView()
    await flushPromises()
    getDashboard.mockClear()

    await wrapper.get('[data-testid="date-range-picker"]').trigger('click')
    await flushPromises()

    expect(getDashboard).toHaveBeenCalledWith({
      start_date: '2026-07-01',
      end_date: '2026-07-15',
    })
  })
})
