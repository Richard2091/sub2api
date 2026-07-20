<template>
  <AppLayout>
    <div class="space-y-4">
      <!-- Profit Summary -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <div v-for="card in statsCards" :key="card.key" class="card p-4">
          <div class="flex items-center gap-3">
            <div :class="['rounded-lg p-2', card.iconBg, card.iconText]">
              <Icon :name="card.icon" size="md" />
            </div>
            <div class="min-w-0">
              <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ card.label }}</p>
              <p :class="['text-xl font-bold tabular-nums', card.valueClass || 'text-gray-900 dark:text-white']">{{ card.value }}</p>
              <p class="truncate text-xs text-gray-500 dark:text-gray-400">{{ card.hint }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Filters -->
      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.dashboard.timeRange') }}:</span>
            <DateRangePicker
              v-model:start-date="startDate"
              v-model:end-date="endDate"
              @change="onDateRangeChange"
            />
          </div>
          <div class="flex-1 sm:max-w-64">
            <input v-model="orderSearch" type="text" :placeholder="t('payment.admin.searchOrders')" class="input" @input="debounceLoadOrders" />
          </div>
          <Select v-model="orderFilters.status" :options="statusFilterOptions" class="w-36" @change="applyFilters" />
          <Select v-model="orderFilters.payment_type" :options="paymentTypeFilterOptions" class="w-40" @change="applyFilters" />
          <Select v-model="orderFilters.order_type" :options="orderTypeFilterOptions" class="w-36" @change="applyFilters" />
          <div class="flex flex-1 flex-wrap items-center justify-end gap-2">
            <button @click="refreshData" :disabled="ordersLoading || statsLoading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="ordersLoading || statsLoading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </div>

      <!-- Analytics Breakdown -->
      <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <div class="card p-4">
          <div class="mb-4 flex items-center justify-between gap-3">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.userTopUpDistribution') }}</h3>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.byPayAmount') }}</span>
          </div>
          <div v-if="!orderStats?.top_users?.length" class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.noData') }}</div>
          <div v-else class="space-y-3">
            <div v-for="user in topUserRows" :key="user.user_id" class="space-y-1.5">
              <div class="flex items-center justify-between gap-3 text-sm">
                <div class="min-w-0">
                  <p class="truncate font-medium text-gray-800 dark:text-gray-100">{{ user.email || user.name || `#${user.user_id}` }}</p>
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ user.order_count }} {{ t('payment.admin.orders') }}</p>
                </div>
                <div class="text-right tabular-nums">
                  <p class="font-semibold text-gray-900 dark:text-white">{{ formatMoney(user.gross_pay_amount) }}</p>
                  <p :class="['text-xs', user.net_pay_amount >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400']">{{ formatMoney(user.net_pay_amount) }}</p>
                </div>
              </div>
              <div class="h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
                <div class="h-full rounded-full bg-blue-500" :style="{ width: `${user.percent}%` }"></div>
              </div>
            </div>
          </div>
        </div>

        <div class="card p-4">
          <div class="mb-4 flex items-center justify-between gap-3">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.subscriptionPurchaseDistribution') }}</h3>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.byRevenue') }}</span>
          </div>
          <div v-if="!orderStats?.subscription_plans?.length" class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.noData') }}</div>
          <div v-else class="overflow-x-auto">
            <table class="min-w-full text-sm">
              <thead>
                <tr class="border-b border-gray-100 text-xs text-gray-500 dark:border-dark-700 dark:text-gray-400">
                  <th class="py-2 text-left font-medium">{{ t('payment.admin.planName') }}</th>
                  <th class="py-2 text-right font-medium">{{ t('payment.admin.orders') }}</th>
                  <th class="py-2 text-right font-medium">{{ t('payment.admin.payAmount') }}</th>
                  <th class="py-2 text-right font-medium">{{ t('payment.admin.netRevenue') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="plan in subscriptionRows" :key="plan.plan_id" class="border-b border-gray-50 last:border-0 dark:border-dark-700/60">
                  <td class="max-w-[12rem] truncate py-2 font-medium text-gray-800 dark:text-gray-100">{{ plan.plan_name }}</td>
                  <td class="py-2 text-right tabular-nums text-gray-600 dark:text-gray-300">{{ plan.order_count }}</td>
                  <td class="py-2 text-right tabular-nums text-gray-900 dark:text-white">{{ formatMoney(plan.gross_pay_amount) }}</td>
                  <td :class="['py-2 text-right tabular-nums font-medium', plan.net_pay_amount >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400']">{{ formatMoney(plan.net_pay_amount) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Table -->
      <OrderTable :orders="orders" :loading="ordersLoading" show-user>
        <template #actions="{ row }">
          <div class="flex items-center gap-1">
            <button @click="showOrderDetail(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-dark-600">
              <Icon name="eye" size="sm" />
              {{ t('common.view') }}
            </button>
            <button v-if="row.status === 'PENDING'" @click="handleCancelOrder(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-yellow-600 hover:bg-yellow-50 dark:text-yellow-400 dark:hover:bg-yellow-900/20">
              <Icon name="x" size="sm" />
              {{ t('payment.orders.cancel') }}
            </button>
            <button v-if="row.status === 'FAILED'" @click="handleRetryOrder(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/20">
              <Icon name="refresh" size="sm" />
              {{ t('payment.admin.retry') }}
            </button>
            <template v-if="row.status === 'REFUND_REQUESTED'">
              <span v-if="row.refund_amount" class="rounded-full bg-purple-100 px-1.5 py-0.5 text-xs font-medium text-purple-700 dark:bg-purple-900/30 dark:text-purple-300">{{ creditedAmountSymbol }}{{ row.refund_amount.toFixed(2) }}</span>
              <button @click="openRefundDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-purple-600 hover:bg-purple-50 dark:text-purple-400 dark:hover:bg-purple-900/20">
                <Icon name="check" size="sm" />
                {{ t('payment.admin.approveRefund') }}
              </button>
            </template>
            <button v-else-if="row.status === 'REFUND_FAILED'" @click="openRefundDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-purple-600 hover:bg-purple-50 dark:text-purple-400 dark:hover:bg-purple-900/20">
              <Icon name="refresh" size="sm" />
              {{ t('payment.admin.retryRefund') }}
            </button>
            <button v-else-if="row.status === 'REFUND_PENDING'" :disabled="refundQueryingIds.has(row.id)" @click="handleQueryRefund(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-orange-600 hover:bg-orange-50 disabled:opacity-60 dark:text-orange-400 dark:hover:bg-orange-900/20">
              <Icon name="refresh" size="sm" :class="refundQueryingIds.has(row.id) ? 'animate-spin' : ''" />
              {{ t('payment.admin.queryRefundStatus') }}
            </button>
            <button v-else-if="row.status === 'COMPLETED' || row.status === 'PARTIALLY_REFUNDED'" @click="openRefundDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20">
              <Icon name="dollar" size="sm" />
              {{ t('payment.admin.refund') }}
            </button>
          </div>
        </template>
      </OrderTable>
      <Pagination v-if="orderPagination.total > 0" :page="orderPagination.page" :total="orderPagination.total" :page-size="orderPagination.page_size" @update:page="handleOrderPageChange" @update:pageSize="handleOrderPageSizeChange" />
    </div>

    <!-- Order Detail Dialog -->
    <BaseDialog :show="showDetailDialog" :title="t('payment.admin.orderDetail')" width="wide" @close="showDetailDialog = false">
      <div v-if="selectedOrder" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</p><p class="font-mono text-sm font-medium text-gray-900 dark:text-white">#{{ selectedOrder.id }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderNo') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedOrder.out_trade_no }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.status') }}</p><OrderStatusBadge :status="selectedOrder.status" /></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.amount') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ creditedAmountSymbol }}{{ selectedOrder.amount.toFixed(2) }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.payAmount') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ paymentAmountSymbol(selectedOrder) }}{{ selectedOrder.pay_amount.toFixed(2) }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.paymentMethod') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ t('payment.methods.' + selectedOrder.payment_type, selectedOrder.payment_type) }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.feeRate') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ selectedOrder.fee_rate }}%</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.createdAt') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(selectedOrder.created_at) }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.expiresAt') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(selectedOrder.expires_at) }}</p></div>
          <div v-if="selectedOrder.paid_at"><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.paidAt') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(selectedOrder.paid_at) }}</p></div>
          <div v-if="selectedOrder.refund_amount"><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundAmount') }}</p><p class="text-sm font-medium text-red-600 dark:text-red-400">{{ creditedAmountSymbol }}{{ selectedOrder.refund_amount.toFixed(2) }}</p></div>
          <div v-if="selectedOrder.refund_reason" class="col-span-2"><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundReason') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ selectedOrder.refund_reason }}</p></div>
          <!-- Refund request info -->
          <div v-if="selectedOrder.refund_requested_at" class="col-span-2 border-t border-gray-200 pt-3 dark:border-dark-600">
            <p class="mb-2 text-xs font-medium text-purple-600 dark:text-purple-400">{{ t('payment.admin.refundRequestInfo') }}</p>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundRequestedAt') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(selectedOrder.refund_requested_at) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundRequestedBy') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">#{{ selectedOrder.refund_requested_by }}</p>
              </div>
              <div class="col-span-2">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundRequestReason') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">{{ selectedOrder.refund_request_reason }}</p>
              </div>
            </div>
          </div>
        </div>
        <!-- Audit Logs -->
        <div v-if="orderAuditLogs.length > 0" class="border-t border-gray-200 pt-4 dark:border-dark-600">
          <p class="mb-2 text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('payment.admin.auditLogs') }}</p>
          <div class="max-h-48 space-y-2 overflow-y-auto">
            <div v-for="log in orderAuditLogs" :key="log.id" class="rounded-lg border border-gray-100 bg-gray-50 p-2.5 dark:border-dark-600 dark:bg-dark-800">
              <div class="flex items-center justify-between">
                <span class="text-xs font-medium text-gray-700 dark:text-gray-300">{{ log.action }}</span>
                <span class="text-xs text-gray-400">{{ formatDateTime(log.created_at) }}</span>
              </div>
              <div v-if="log.detail" class="mt-1 break-all text-xs text-gray-500 dark:text-gray-400">{{ log.detail }}</div>
              <div v-if="log.operator" class="mt-1 text-xs text-gray-400">{{ t('payment.admin.operator') }}: {{ log.operator }}</div>
            </div>
          </div>
        </div>
      </div>
    </BaseDialog>

    <AdminRefundDialog :show="showRefundDialog" :order="selectedOrder" :submitting="refundSubmitting" @confirm="handleRefund" @cancel="showRefundDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatOrderDateTime } from '@/components/payment/orderUtils'
import type { AdminOrderStats, PaymentOrder } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Icon from '@/components/icons/Icon.vue'
import AdminRefundDialog from '@/components/admin/payment/AdminRefundDialog.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import OrderTable from '@/components/payment/OrderTable.vue'
import { currencySymbol } from '@/components/payment/currency'

interface AuditLog {
  id: number
  action: string
  detail: string | null
  operator: string | null
  created_at: string
}

const { t } = useI18n()
const appStore = useAppStore()

const formatLD = (d: Date) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
const getLast24HoursRangeDates = (): { start: string; end: string } => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return { start: formatLD(start), end: formatLD(end) }
}
const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)
const ordersLoading = ref(false)
const statsLoading = ref(false)
const orders = ref<PaymentOrder[]>([])
const orderStats = ref<AdminOrderStats | null>(null)
const orderSearch = ref('')
const orderFilters = reactive({ status: '', payment_type: '', order_type: '' })
const orderPagination = reactive({ page: 1, page_size: 20, total: 0 })
const selectedOrder = ref<PaymentOrder | null>(null)
const showDetailDialog = ref(false)
const showRefundDialog = ref(false)
const refundSubmitting = ref(false)
const refundQueryingIds = ref(new Set<number>())
const orderAuditLogs = ref<AuditLog[]>([])
const creditedAmountSymbol = currencySymbol('USD')

const moneyFormatter = new Intl.NumberFormat(undefined, {
  style: 'currency',
  currency: 'USD',
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})

function formatMoney(value: number | null | undefined): string {
  return moneyFormatter.format(Number(value || 0))
}

function buildOrderQueryParams() {
  return {
    keyword: orderSearch.value || undefined,
    status: orderFilters.status || undefined,
    payment_type: orderFilters.payment_type || undefined,
    order_type: orderFilters.order_type || undefined,
    start_date: startDate.value || undefined,
    end_date: endDate.value || undefined,
  }
}

const statsCards = computed(() => {
  const stats = orderStats.value
  return [
    {
      key: 'net',
      icon: 'dollar' as const,
      iconBg: 'bg-green-100 dark:bg-green-900/30',
      iconText: 'text-green-600 dark:text-green-400',
      label: t('payment.admin.netRevenue'),
      value: formatMoney(stats?.net_pay_amount),
      hint: `${t('payment.admin.payAmount')} ${formatMoney(stats?.gross_pay_amount)} / ${t('payment.admin.refundAmount')} ${formatMoney(stats?.refund_amount)}`,
      valueClass: (stats?.net_pay_amount || 0) >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400',
    },
    {
      key: 'amount',
      icon: 'creditCard' as const,
      iconBg: 'bg-blue-100 dark:bg-blue-900/30',
      iconText: 'text-blue-600 dark:text-blue-400',
      label: t('payment.admin.grossRevenue'),
      value: formatMoney(stats?.gross_amount),
      hint: `${t('payment.admin.payAmount')} ${formatMoney(stats?.gross_pay_amount)}`,
    },
    {
      key: 'fee',
      icon: 'calculator' as const,
      iconBg: 'bg-amber-100 dark:bg-amber-900/30',
      iconText: 'text-amber-600 dark:text-amber-400',
      label: t('payment.admin.estimatedFee'),
      value: formatMoney(stats?.fee_amount),
      hint: `${t('payment.admin.avgAmount')} ${formatMoney(stats?.avg_pay_amount)}`,
    },
    {
      key: 'orders',
      icon: 'document' as const,
      iconBg: 'bg-purple-100 dark:bg-purple-900/30',
      iconText: 'text-purple-600 dark:text-purple-400',
      label: t('payment.admin.orderCount'),
      value: String(stats?.total_orders || 0),
      hint: `${t('payment.admin.paidOrders')} ${stats?.paid_orders || 0} / ${t('payment.status.pending')} ${stats?.pending_orders || 0}`,
    },
  ]
})

const topUserRows = computed(() => {
  const rows = orderStats.value?.top_users || []
  const max = Math.max(...rows.map((item) => item.gross_pay_amount), 0)
  return rows.slice(0, 8).map((item) => ({
    ...item,
    percent: max > 0 ? Math.max(4, Math.round((item.gross_pay_amount / max) * 100)) : 0,
  }))
})

const subscriptionRows = computed(() => (orderStats.value?.subscription_plans || []).slice(0, 8))

function paymentAmountSymbol(order: PaymentOrder | null | undefined): string {
  return currencySymbol(order?.currency)
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function debounceLoadOrders() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { orderPagination.page = 1; refreshData() }, 300)
}

async function loadOrders() {
  ordersLoading.value = true
  try {
    const res = await adminPaymentAPI.getOrders({
      page: orderPagination.page, page_size: orderPagination.page_size,
      ...buildOrderQueryParams(),
    })
    orders.value = res.data.items || []
    orderPagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally { ordersLoading.value = false }
}

async function loadOrderStats() {
  statsLoading.value = true
  try {
    const res = await adminPaymentAPI.getOrderStats(buildOrderQueryParams())
    orderStats.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally { statsLoading.value = false }
}

function refreshData() {
  loadOrders()
  loadOrderStats()
}

function applyFilters() {
  orderPagination.page = 1
  refreshData()
}

function onDateRangeChange(range: { startDate: string; endDate: string; preset: string | null }) {
  startDate.value = range.startDate
  endDate.value = range.endDate
  orderPagination.page = 1
  refreshData()
}

function handleOrderPageChange(page: number) { orderPagination.page = page; loadOrders() }
function handleOrderPageSizeChange(size: number) { orderPagination.page_size = size; orderPagination.page = 1; loadOrders() }

const statusFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allStatuses') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'PAID', label: t('payment.status.paid') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'EXPIRED', label: t('payment.status.expired') },
  { value: 'CANCELLED', label: t('payment.status.cancelled') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
  { value: 'REFUND_REQUESTED', label: t('payment.status.refund_requested') },
  { value: 'REFUND_PENDING', label: t('payment.status.refund_pending') },
  { value: 'REFUND_FAILED', label: t('payment.status.refund_failed') },
])

const paymentTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allPaymentTypes') },
  { value: 'alipay', label: t('payment.methods.alipay') },
  { value: 'wxpay', label: t('payment.methods.wxpay') },
  { value: 'stripe', label: t('payment.methods.stripe') },
  { value: 'airwallex', label: t('payment.methods.airwallex') },
])

const orderTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allOrderTypes') },
  { value: 'balance', label: t('payment.admin.balanceOrder') },
  { value: 'subscription', label: t('payment.admin.subscriptionOrder') },
])

async function showOrderDetail(order: PaymentOrder) {
  selectedOrder.value = order
  orderAuditLogs.value = []
  showDetailDialog.value = true
  try {
    const res = await adminPaymentAPI.getOrder(order.id)
    const data = res.data as unknown as Record<string, unknown>
    if (data.order) selectedOrder.value = data.order as PaymentOrder
    orderAuditLogs.value = ((data.auditLogs || data.audit_logs || []) as unknown) as AuditLog[]
  } catch (_err: unknown) { /* keep cached order data */ }
}

async function handleCancelOrder(order: PaymentOrder) {
  try { await adminPaymentAPI.cancelOrder(order.id); appStore.showSuccess(t('payment.admin.orderCancelled')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

async function handleRetryOrder(order: PaymentOrder) {
  try { await adminPaymentAPI.retryRecharge(order.id); appStore.showSuccess(t('payment.admin.retrySuccess')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

function openRefundDialog(order: PaymentOrder) { selectedOrder.value = order; showRefundDialog.value = true }

function isRefundPendingWarning(warning: string | undefined): boolean {
  return /pending|处理中|待/.test(String(warning || '').toLowerCase())
}

async function handleRefund(data: { amount: number; reason: string; deduct_balance: boolean; force: boolean }) {
  if (!selectedOrder.value) return
  refundSubmitting.value = true
  try {
    const res = await adminPaymentAPI.refundOrder(selectedOrder.value.id, { amount: data.amount, reason: data.reason, deduct_balance: data.deduct_balance, force: data.force })
    if (res.data.success) {
      appStore.showSuccess(t('payment.admin.refundSuccess'))
      showRefundDialog.value = false
      loadOrders()
      return
    }
    if (isRefundPendingWarning(res.data.warning)) {
      appStore.showSuccess(t('payment.admin.refundPending'))
      showRefundDialog.value = false
      loadOrders()
      return
    }
    appStore.showError(res.data.warning || t('common.error'))
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { refundSubmitting.value = false }
}

async function handleQueryRefund(order: PaymentOrder) {
  refundQueryingIds.value = new Set(refundQueryingIds.value).add(order.id)
  try {
    const res = await adminPaymentAPI.queryRefund(order.id)
    if (res.data.success) {
      appStore.showSuccess(t('payment.admin.refundSuccess'))
    } else if (isRefundPendingWarning(res.data.warning)) {
      appStore.showSuccess(t('payment.admin.refundPending'))
    } else {
      appStore.showError(res.data.warning || t('common.error'))
    }
    loadOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    const next = new Set(refundQueryingIds.value)
    next.delete(order.id)
    refundQueryingIds.value = next
  }
}

function formatDateTime(dateStr: string): string { return formatOrderDateTime(dateStr) }

onMounted(() => refreshData())
</script>
