<script>
  let { size = 32 } = $props();
</script>

<!-- The same drawing as public/icon.svg, with the theme's line colours. -->
<svg width={size} height={size} viewBox="0 0 64 64" aria-hidden="true">
  <rect width="64" height="64" rx="14" fill="#17181a" />
  <g fill="none" stroke-width="7.5">
    <path d="M42.96 21.04A15.5 15.5 0 0 0 26.95 17.34" stroke="var(--blue)" />
    <path d="M25.21 18.07A15.5 15.5 0 0 0 16.53 31.05" stroke="var(--yellow)" />
    <path d="M16.53 32.95A15.5 15.5 0 0 0 25.21 45.93" stroke="var(--green)" />
    <path d="M26.95 46.66A15.5 15.5 0 0 0 42.96 42.96" stroke="var(--red)" />
  </g>
  <circle cx="42.96" cy="21.04" r="4.4" fill="#fff" stroke="#17181a" stroke-width="2.4" />
  <circle cx="42.96" cy="42.96" r="4.4" fill="#fff" stroke="#17181a" stroke-width="2.4" />
</svg>
