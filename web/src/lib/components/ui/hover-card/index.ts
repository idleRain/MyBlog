import { LinkPreview as HoverCardPrimitive } from 'bits-ui'
import Trigger from './hover-card-trigger.svelte'
import Content from './hover-card-content.svelte'

const Root = HoverCardPrimitive.Root

export {
  Root,
  Content,
  Trigger,
  Root as HoverCard,
  Content as HoverCardContent,
  Trigger as HoverCardTrigger
}
