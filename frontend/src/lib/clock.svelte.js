// One shared ticking clock, so every countdown on screen moves in step.
class Clock {
  now = $state(Date.now());

  constructor() {
    setInterval(() => (this.now = Date.now()), 1000);
  }
}

export const clock = new Clock();
