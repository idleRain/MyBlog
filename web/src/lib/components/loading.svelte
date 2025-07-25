<script lang="ts">
</script>

<div class="fixed inset-0 flex items-center justify-center bg-[#0002]">
  <div class="loader"></div>
</div>

<style>
.loader {
  display: grid;
  aspect-ratio: 1;
  width: 50px;
}
.loader::before,
.loader::after {
  grid-area: 1/1;
  content: '';
  --c: no-repeat radial-gradient(farthest-side, #25b09b 92%, #0000);
  animation: l12 1s infinite;
  background:
    var(--c) 50% 0,
    var(--c) 50% 100%,
    var(--c) 100% 50%,
    var(--c) 0 50%;
  background-size: 12px 12px;
}
.loader::before {
  filter: hue-rotate(45deg);
  animation-timing-function: linear;
  margin: 4px;
  background-size: 8px 8px;
}

@keyframes l12 {
  100% {
    transform: rotate(0.5turn);
  }
}
</style>
