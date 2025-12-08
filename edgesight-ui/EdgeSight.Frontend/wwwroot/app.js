const statusEl = document.getElementById('status');
const answerSection = document.getElementById('answerSection');
const answerEl = document.getElementById('answer');
const sourcesEl = document.getElementById('sources');
const btn = document.getElementById('askBtn');

btn.addEventListener('click', submitQuery);

async function submitQuery() {
  const q = document.getElementById('question').value.trim();
  const location = document.getElementById('location').value.trim() || 'Los Angeles';
  if (!q) {
    statusEl.textContent = 'Enter a question first.';
    return;
  }

  setBusy(true);
  statusEl.textContent = 'Thinking...';
  answerSection.style.display = 'none';

  try {
    const resp = await fetch(`/api/query?q=${encodeURIComponent(q)}&location=${encodeURIComponent(location)}`);
    if (!resp.ok) {
      throw new Error(`API ${resp.status}`);
    }
    const payload = await resp.json();
    renderResult(payload);
    statusEl.textContent = 'Done';
  } catch (err) {
    statusEl.textContent = `Error: ${err.message}`;
  } finally {
    setBusy(false);
  }
}

function renderResult(payload) {
  const answer = payload.answer || 'No answer available.';
  answerEl.textContent = answer;

  sourcesEl.innerHTML = '';
  if (Array.isArray(payload.sources)) {
    payload.sources.forEach((s, idx) => {
      const card = document.createElement('div');
      card.className = 'source-card';
      const ts = s.snapshot_ts || 'unknown time';
      const loc = s.location || 'unknown location';
      card.innerHTML = `
        <h4>Source ${idx + 1}</h4>
        <div class="meta">${ts} · ${loc} · score ${(s.score ?? 0).toFixed(3)}</div>
        <div>${s.summary || ''}</div>
      `;
      sourcesEl.appendChild(card);
    });
  }
  answerSection.style.display = 'block';
}

function setBusy(isBusy) {
  btn.disabled = isBusy;
  btn.textContent = isBusy ? 'Working...' : 'Ask';
}
