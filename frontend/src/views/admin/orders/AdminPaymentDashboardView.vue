<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header with date range filter -->
      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-4">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.dashboard.timeRange') }}:</span>
            <DateRangePicker
              v-model:start-date="startDate"
              v-model:end-date="endDate"
              @change="onDateRangeChange"
            />
          </div>
          <div class="ml-auto flex items-center gap-2">
            <button @click="loadDashboard" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </div>

      <!-- Dashboard Content -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>
      <template v-else-if="stats">
        <OrderStatsCards :stats="stats" />
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <div v-for="card in profitCards" :key="card.key" class="card p-4">
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
        <DailyRevenueChart :data="stats.daily_series || []" :loading="loading" />
        <div class="grid grid-cols-1 gap-6 xl:grid-cols-3">
          <div class="card p-4">
            <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.paymentDistribution') }}</h3>
            <div v-if="!stats.payment_methods?.length" class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.noData') }}</div>
            <div v-else class="space-y-3">
              <div v-for="method in stats.payment_methods" :key="method.type" class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <span :class="['inline-block h-3 w-3 rounded-full', methodColor(method.type)]"></span>
                  <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('payment.methods.' + method.type, method.type) }}</span>
                </div>
                <div class="text-right">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">{{ formatMoney(method.amount) }}</span>
                  <span class="ml-2 text-xs text-gray-500 dark:text-gray-400">({{ method.count }})</span>
                </div>
              </div>
            </div>
          </div>
          <div class="card p-4">
            <div class="mb-4 flex items-center justify-between gap-3">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.userTopUpDistribution') }}</h3>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.byPayAmount') }}</span>
            </div>
            <div v-if="!profitStats?.top_users?.length" class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.noData') }}</div>
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
            <div v-if="!profitStats?.subscription_plans?.length" class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.noData') }}</div>
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
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { DashboardStats } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderStatsCards from '@/components/admin/payment/OrderStatsCards.vue'
import DailyRevenueChart from '@/components/admin/payment/DailyRevenueChart.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const stats = ref<DashboardStats | null>(null)

const formatLD = (d: Date) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const getLast24HoursRangeDates = (): { start: string; end: string } => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return {
    start: formatLD(start),
    end: formatLD(end)
  }
}

const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)

function methodColor(type: string): string {
  const c: Record<string, string> = {
    alipay: 'bg-blue-500', wxpay: 'bg-green-500',
    alipay_direct: 'bg-blue-400', wxpay_direct: 'bg-green-400',
    stripe: 'bg-purple-500',
  }
  return c[type] || 'bg-gray-400'
}

const moneyFormatter = new Intl.NumberFormat(undefined, {
  style: 'currency',
  currency: 'USD',
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})

function formatMoney(value: number | null | undefined): string {
  return moneyFormatter.format(Number(value || 0))
}

const profitStats = computed(() => stats.value?.profit_stats || null)

const profitCards = computed(() => {
  const profit = profitStats.value
  return [
    {
      key: 'net',
      icon: 'dollar' as const,
      iconBg: 'bg-green-100 dark:bg-green-900/30',
      iconText: 'text-green-600 dark:text-green-400',
      label: t('payment.admin.netRevenue'),
      value: formatMoney(profit?.net_pay_amount),
      hint: `${t('payment.admin.payAmount')} ${formatMoney(profit?.gross_pay_amount)} / ${t('payment.admin.refundAmount')} ${formatMoney(profit?.refund_amount)}`,
      valueClass: (profit?.net_pay_amount || 0) >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400',
    },
    {
      key: 'amount',
      icon: 'creditCard' as const,
      iconBg: 'bg-blue-100 dark:bg-blue-900/30',
      iconText: 'text-blue-600 dark:text-blue-400',
      label: t('payment.admin.grossRevenue'),
      value: formatMoney(profit?.gross_amount),
      hint: `${t('payment.admin.payAmount')} ${formatMoney(profit?.gross_pay_amount)}`,
    },
    {
      key: 'fee',
      icon: 'calculator' as const,
      iconBg: 'bg-amber-100 dark:bg-amber-900/30',
      iconText: 'text-amber-600 dark:text-amber-400',
      label: t('payment.admin.estimatedFee'),
      value: formatMoney(profit?.fee_amount),
      hint: `${t('payment.admin.avgAmount')} ${formatMoney(profit?.avg_pay_amount)}`,
    },
    {
      key: 'orders',
      icon: 'document' as const,
      iconBg: 'bg-purple-100 dark:bg-purple-900/30',
      iconText: 'text-purple-600 dark:text-purple-400',
      label: t('payment.admin.orderCount'),
      value: String(profit?.total_orders || 0),
      hint: `${t('payment.admin.paidOrders')} ${profit?.paid_orders || 0} / ${t('payment.status.pending')} ${profit?.pending_orders || 0}`,
    },
  ]
})

const topUserRows = computed(() => {
  const rows = profitStats.value?.top_users || []
  const max = Math.max(...rows.map((item) => item.gross_pay_amount), 0)
  return rows.slice(0, 8).map((item) => ({
    ...item,
    percent: max > 0 ? Math.max(4, Math.round((item.gross_pay_amount / max) * 100)) : 0,
  }))
})

const subscriptionRows = computed(() => (profitStats.value?.subscription_plans || []).slice(0, 8))

async function loadDashboard() {
  loading.value = true
  try {
    const res = await adminPaymentAPI.getDashboard({
      start_date: startDate.value,
      end_date: endDate.value,
    })
    stats.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function onDateRangeChange(range: { startDate: string; endDate: string; preset: string | null }) {
  startDate.value = range.startDate
  endDate.value = range.endDate
  loadDashboard()
}

onMounted(() => loadDashboard())
</script>
