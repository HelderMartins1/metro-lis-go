import { mount } from 'svelte';
import { registerSW } from 'virtual:pwa-register';
import '@fontsource/space-grotesk/400.css';
import '@fontsource/space-grotesk/500.css';
import '@fontsource/space-grotesk/600.css';
import '@fontsource/space-grotesk/700.css';
import './app.css';
import App from './App.svelte';

registerSW({ immediate: true });

mount(App, { target: document.getElementById('app') });
