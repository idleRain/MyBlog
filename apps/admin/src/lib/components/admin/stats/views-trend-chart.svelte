<script lang="ts">
interface Props {
  dates: string[]
  values: number[]
  height?: number
}

let { dates, values, height = 220 }: Props = $props()

const width = 600
const padding = 20

/**
 * 计算折线各点坐标，值为零时仍保留基线位置。
 */
const points = $derived.by(() => {
  if (values.length === 0) return ''
  const max = Math.max(1, ...values)
  const stepX = (width - padding * 2) / Math.max(1, values.length - 1)
  return values
    .map((value, index) => {
      const x = padding + index * stepX
      const y = height - padding - (value / max) * (height - padding * 2)
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
})

// 面积填充路径，从折线首尾闭合到图表底部。
const areaPath = $derived.by(() => {
  if (points === '') return ''
  const firstX = padding
  const lastX = width - padding
  return `M ${firstX} ${height - padding} L ${points.replaceAll(' ', ' L ')} L ${lastX} ${height - padding} Z`
})
</script>

{#if values.length > 0}
  <svg
    viewBox={`0 0 ${width} ${height}`}
    class="h-full w-full"
    preserveAspectRatio="none"
    role="img"
    aria-label="文章浏览量趋势图"
  >
    <defs>
      <linearGradient id="trend-area" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="var(--color-chart-1)" stop-opacity="0.25" />
        <stop offset="100%" stop-color="var(--color-chart-1)" stop-opacity="0" />
      </linearGradient>
    </defs>
    <path d={areaPath} fill="url(#trend-area)" />
    <polyline
      {points}
      fill="none"
      stroke="var(--color-chart-1)"
      stroke-width="2"
      stroke-linejoin="round"
      stroke-linecap="round"
    />
  </svg>

  <div class="mt-2 flex justify-between text-xs text-muted-foreground">
    <span>{dates[0]}</span>
    <span>{dates[dates.length - 1]}</span>
  </div>
{:else}
  <div class="flex h-48 items-center justify-center text-sm text-muted-foreground">
    暂无趋势数据
  </div>
{/if}
