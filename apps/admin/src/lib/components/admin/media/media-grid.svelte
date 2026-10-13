<script lang="ts">
import { MEDIA_IMAGE_MIME_PREFIX, MEDIA_VIDEO_MIME_PREFIX } from '$lib/constants/media'
import { Copy, FileText, Film, Image, Trash2 } from '@lucide/svelte'
import type { MediaFile } from '@myblog/api/modules/media/types'
import { getFileSize } from '@myblog/shared'
import { Button } from '$ui/button'

interface Props {
  media: MediaFile[]
  onCopy: (file: MediaFile) => void
  onDelete: (file: MediaFile) => void
}

let { media, onCopy, onDelete }: Props = $props()

// 依据 MIME 前缀返回文件类型图标，图片走缩略图预览时不使用该图标。
function typeIcon(file: MediaFile) {
  if (file.mimeType.startsWith(MEDIA_IMAGE_MIME_PREFIX)) return Image
  if (file.mimeType.startsWith(MEDIA_VIDEO_MIME_PREFIX)) return Film
  return FileText
}
</script>

<!-- 媒体网格：窄屏两列，lg 与 xl 断点依次扩展到三列与四列 -->
<div class="grid grid-cols-2 gap-4 p-4 lg:grid-cols-3 xl:grid-cols-4">
  {#each media as file (file.id)}
    <div class="overflow-hidden rounded-lg border bg-card">
      <div class="flex h-36 items-center justify-center border-b bg-muted/40">
        {#if file.mimeType.startsWith(MEDIA_IMAGE_MIME_PREFIX)}
          <img
            src={file.thumbnailUrl || file.fileUrl}
            alt={file.filename}
            loading="lazy"
            class="h-full w-full object-cover"
          />
        {:else}
          {@const IconComponent = typeIcon(file)}
          <IconComponent class="size-10 text-muted-foreground" />
        {/if}
      </div>
      <div class="flex flex-col gap-2 p-3">
        <p class="truncate text-sm font-medium" title={file.filename}>{file.filename}</p>
        <p class="truncate text-xs text-muted-foreground">
          {getFileSize(file.fileSize)} · {file.mimeType}
        </p>
        <div class="flex gap-2">
          <Button variant="outline" size="sm" class="flex-1" onclick={() => onCopy(file)}>
            <Copy data-icon="inline-start" />
            复制链接
          </Button>
          <Button
            variant="outline"
            size="icon"
            class="text-destructive hover:text-destructive"
            aria-label={`删除文件 ${file.filename}`}
            onclick={() => onDelete(file)}
          >
            <Trash2 />
          </Button>
        </div>
      </div>
    </div>
  {/each}
</div>
