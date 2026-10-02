<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { format } from 'date-fns'
import api from '@/api'
import { Button } from '@shared-ui/components/ui/button'
import { Badge } from '@shared-ui/components/ui/badge'
import { Input } from '@shared-ui/components/ui/input'
import { Label } from '@shared-ui/components/ui/label'
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem
} from '@shared-ui/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@shared-ui/components/ui/table'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger
} from '@shared-ui/components/ui/collapsible'
import { handleHTTPError } from '@shared-ui/utils/http'
import { ChevronDown } from 'lucide-vue-next'
import {
  DELIVERY_STATES,
  DELIVERY_STATE_KEYS,
  deliveryPage,
  deliveryState,
  deliveryWho
} from './campaignHistory.js'

const PAGE_SIZE = 50
const ANY = 'any'
const STATE_VARIANTS = {
  replied: 'default',
  opened: 'secondary',
  dismissed: 'outline',
  displayed: 'outline',
  undisplayed: 'outline'
}

const props = defineProps({
  inboxId: { type: Number, default: 0 },
  campaigns: { type: Array, default: () => [] }
})
const { t } = useI18n()

const rows = ref([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const error = ref('')
const from = ref(new Date(Date.now() - 7 * 86400000).toISOString().slice(0, 10))
const to = ref(new Date().toISOString().slice(0, 10))
const state = ref(ANY)
const campaignId = ref(ANY)
const url = ref('')

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))
const campaignName = (id) =>
  props.campaigns.find((item) => item.id === id)?.name || t('widget.deletedCampaign')

const load = async (target = 1) => {
  if (!props.inboxId) return
  loading.value = true
  error.value = ''
  try {
    const params = { from: from.value, to: to.value, page: target, page_size: PAGE_SIZE }
    if (state.value !== ANY) params.state = state.value
    if (campaignId.value !== ANY) params.campaign_id = campaignId.value
    if (url.value.trim()) params.url = url.value.trim()
    const data = (await api.getCampaignDeliveries(props.inboxId, params)).data.data
    rows.value = data.results || []
    total.value = data.total || 0
    page.value = data.page || target
  } catch (err) {
    error.value = handleHTTPError(err).message
  } finally {
    loading.value = false
  }
}

onMounted(() => load())
</script>

<template>
  <Collapsible v-if="inboxId" class="space-y-4" data-testid="campaign-history">
    <CollapsibleTrigger
      type="button"
      class="flex items-center gap-2 text-base font-semibold text-foreground group"
    >
      <ChevronDown class="size-4 transition-transform group-data-[state=closed]:-rotate-90" />
      {{ t('widget.campaignHistory.title') }}
    </CollapsibleTrigger>
    <CollapsibleContent class="space-y-4">
      <p class="text-xs text-muted-foreground">{{ t('widget.campaignHistory.hint') }}</p>
      <p v-if="error" role="alert" class="text-sm text-destructive">{{ error }}</p>
      <div class="flex flex-wrap gap-3 items-end">
        <div class="space-y-2">
          <Label for="campaign-history-from">{{ t('globals.terms.from') }}</Label>
          <Input id="campaign-history-from" v-model="from" type="date" />
        </div>
        <div class="space-y-2">
          <Label for="campaign-history-to">{{ t('globals.terms.to') }}</Label>
          <Input id="campaign-history-to" v-model="to" type="date" />
        </div>
        <div class="space-y-2">
          <Label for="campaign-history-campaign">{{ t('globals.terms.message', 1) }}</Label>
          <Select v-model="campaignId">
            <SelectTrigger id="campaign-history-campaign" class="w-48">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="ANY">{{ t('globals.messages.all') }}</SelectItem>
              <SelectItem v-for="item in campaigns" :key="item.id" :value="item.id">
                {{ item.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="space-y-2">
          <Label for="campaign-history-state">{{ t('globals.terms.status', 1) }}</Label>
          <Select v-model="state">
            <SelectTrigger id="campaign-history-state" class="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="ANY">{{ t('globals.messages.all') }}</SelectItem>
              <SelectItem v-for="key in DELIVERY_STATES" :key="key" :value="key">
                {{ t(DELIVERY_STATE_KEYS[key]) }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="space-y-2">
          <Label for="campaign-history-url">{{ t('widget.campaignHistory.pageContains') }}</Label>
          <Input id="campaign-history-url" v-model="url" class="w-48" placeholder="/pricing" />
        </div>
        <Button type="button" variant="outline" :disabled="loading" @click="load(1)">
          {{ t('globals.terms.refresh') }}
        </Button>
      </div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{{ t('widget.campaignHistory.sentAt') }}</TableHead>
            <TableHead>{{ t('globals.terms.page', 1) }}</TableHead>
            <TableHead>{{ t('globals.terms.message', 1) }}</TableHead>
            <TableHead>{{ t('widget.campaignHistory.sentTo') }}</TableHead>
            <TableHead>{{ t('globals.terms.status', 1) }}</TableHead>
            <TableHead>{{ t('globals.terms.conversation', 1) }}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="row in rows" :key="row.id">
            <TableCell class="whitespace-nowrap tabular-nums">
              {{ format(new Date(row.created_at), 'yyyy-MM-dd HH:mm') }}
            </TableCell>
            <TableCell class="max-w-56 truncate" :title="row.url">
              {{ deliveryPage(row.url) || '-' }}
            </TableCell>
            <TableCell class="max-w-80">
              <div class="whitespace-pre-line">{{ row.message }}</div>
              <div class="text-xs text-muted-foreground">{{ campaignName(row.campaign_id) }}</div>
            </TableCell>
            <TableCell class="whitespace-nowrap">
              <div>{{ deliveryWho(row, t('globals.terms.visitor', 1)) }}</div>
              <div class="text-xs text-muted-foreground">
                {{ row.mobile ? t('globals.terms.mobile') : t('globals.terms.desktop') }}
              </div>
            </TableCell>
            <TableCell>
              <Badge :variant="STATE_VARIANTS[deliveryState(row)]">
                {{ t(DELIVERY_STATE_KEYS[deliveryState(row)]) }}
              </Badge>
            </TableCell>
            <TableCell>
              <RouterLink
                v-if="row.conversation_uuid"
                class="text-primary underline underline-offset-2"
                :to="{ name: 'inbox-conversation', params: { type: 'all', uuid: row.conversation_uuid } }"
              >
                {{ t('widget.campaignHistory.openConversation') }}
              </RouterLink>
              <span v-else class="text-muted-foreground">-</span>
            </TableCell>
          </TableRow>
          <TableRow v-if="!rows.length">
            <TableCell :colspan="6" class="text-muted-foreground">
              {{ t('globals.messages.noResultsFound') }}
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
      <div class="flex items-center justify-between text-sm text-muted-foreground">
        <span>{{ t('widget.campaignHistory.total', { count: total }) }}</span>
        <div class="flex items-center gap-2">
          <Button type="button" variant="outline" size="sm" :disabled="loading || page <= 1" @click="load(page - 1)">
            {{ t('widget.campaignHistory.previous') }}
          </Button>
          <span class="tabular-nums">{{ page }} / {{ totalPages }}</span>
          <Button type="button" variant="outline" size="sm" :disabled="loading || page >= totalPages" @click="load(page + 1)">
            {{ t('widget.campaignHistory.next') }}
          </Button>
        </div>
      </div>
    </CollapsibleContent>
  </Collapsible>
</template>
