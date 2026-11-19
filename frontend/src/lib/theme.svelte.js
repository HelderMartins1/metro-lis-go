// The look: 'auto' follows the device, 'light' and 'dark' override it.
// index.html applies the saved choice before the first paint; this keeps it in step.

const KEY = 'cais:theme';
const COLORS = { light: '#f6f6f3', dark: '#101113' };
const media = matchMedia('(prefers-color-scheme: dark)');

function read() {
  try {
    const value = JSON.parse(localStorage.getItem(KEY));
    return ['auto', 'light', 'dark'].includes(value) ? value : 'auto';
  } catch {
    return 'auto';
  }
}

class Theme {
  choice = $state(read());
  system = $state(media.matches ? 'dark' : 'light');
  resolved = $derived(this.choice === 'auto' ? this.system : this.choice);

  constructor() {
    media.addEventListener('change', (event) => (this.system = event.matches ? 'dark' : 'light'));
    $effect.root(() => {
      $effect(() => {
        document.documentElement.dataset.theme = this.resolved;
        document.querySelector('meta[name=theme-color]')?.setAttribute('content', COLORS[this.resolved]);
      });
    });
  }

  set(choice) {
    this.choice = choice;
    try {
      localStorage.setItem(KEY, JSON.stringify(choice));
    } catch {}
  }
}

export const theme = new Theme();
