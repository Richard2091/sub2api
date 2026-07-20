import { describe, expect, it, beforeEach, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AdminOrdersView from '../AdminOrdersView.vue'

const getOrders = vi.hoisted(() => vi.fn())

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
    getOrder: vi.fn(),
    cancelOrder: vi.fn(),
    retryRecharge: vi.fn(),
    refundOrder: vi.fn(),
    queryRefund: vi.fn(),
  },
  default: {
    getOrders,
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

describe('AdminOrdersView filters', () => {
  beforeEach(() => {
    getOrders.mockResolvedValue({ data: { items: [], total: 0 } })
  })

  it('does not render the payment overview profit analytics', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).not.toContain('$120.43')
    expect(wrapper.text()).not.toContain('alice@example.com')
    expect(wrapper.text()).not.toContain('Pro')
  })

  it('reloads orders with the selected date range', async () => {
    const wrapper = mountView()
    await flushPromises()

    getOrders.mockClear()
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
  })
})
