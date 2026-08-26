// push.js — handles clicks on `.push-btn`. Hits POST /push/device.
//
// Defaults to dry-run. Confirm() asks the operator one more time before
// apply, because a misfire is the worst class of bug in this codebase.

document.addEventListener('click', async (ev) => {
  const btn = ev.target.closest('.push-btn');
  if (!btn || btn.disabled) return;
  const deviceId = btn.dataset.deviceId;

  let apply = false;
  if (btn.classList.contains('apply')) apply = true;

  if (apply) {
    const ok = window.confirm(`Apply push to ${deviceId}? This mutates the running configuration.`);
    if (!ok) return;
  }

  btn.disabled = true;
  const status = btn.parentElement.querySelector('.push-status') || (() => {
    const s = document.createElement('span');
    s.className = 'push-status';
    btn.parentElement.appendChild(s);
    return s;
  })();
  status.textContent = 'pushing…';

  try {
    const r = await fetch('/push/device', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ device_id: deviceId, apply }),
    });
    const body = await r.json();
    status.textContent = body.status || (r.ok ? 'ok' : 'failed');
  } catch (err) {
    status.textContent = `error: ${err}`;
  } finally {
    btn.disabled = false;
  }
});
