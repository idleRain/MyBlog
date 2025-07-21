import { AlertDialog as AlertDialogPrimitive } from 'bits-ui'
import Description from './alert-dialog-description.svelte'
import Trigger from './alert-dialog-trigger.svelte'
import Overlay from './alert-dialog-overlay.svelte'
import Content from './alert-dialog-content.svelte'
import Header from './alert-dialog-header.svelte'
import Footer from './alert-dialog-footer.svelte'
import Cancel from './alert-dialog-cancel.svelte'
import Action from './alert-dialog-action.svelte'
import Title from './alert-dialog-title.svelte'

const Root = AlertDialogPrimitive.Root
const Portal = AlertDialogPrimitive.Portal

export {
  Root,
  Title,
  Action,
  Cancel,
  Portal,
  Footer,
  Header,
  Trigger,
  Overlay,
  Content,
  Description,
  //
  Root as AlertDialog,
  Title as AlertDialogTitle,
  Action as AlertDialogAction,
  Cancel as AlertDialogCancel,
  Portal as AlertDialogPortal,
  Footer as AlertDialogFooter,
  Header as AlertDialogHeader,
  Trigger as AlertDialogTrigger,
  Overlay as AlertDialogOverlay,
  Content as AlertDialogContent,
  Description as AlertDialogDescription
}
