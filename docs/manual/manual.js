/* Interacciones del manual. Ninguna ejecuta comandos, agentes ni solicitudes de red. */
(() => {
  'use strict';
  const scenarios = window.LufyManualScenarios;
  if (!Array.isArray(scenarios) || !scenarios.length) return;
  const byId = (id) => document.getElementById(id);
  const state = { scenario: 0, step: 0 };
  const tabs = byId('scenario-tabs');
  const promptKind = byId('prompt-kind');
  const journey = byId('journey');
  const setText = (id, text) => { byId(id).textContent = text; };
  const current = () => scenarios[state.scenario];

  scenarios.forEach((scenario, index) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'scenario-tab';
    button.id = `scenario-tab-${scenario.id}`;
    button.setAttribute('role', 'tab');
    button.setAttribute('aria-controls', 'scenario-panel');
    const number = document.createElement('span');
    number.textContent = String(index + 1).padStart(2, '0');
    number.setAttribute('aria-hidden', 'true');
    button.append(number, document.createTextNode(scenario.label));
    button.addEventListener('click', () => chooseScenario(index));
    button.addEventListener('keydown', (event) => {
      let target;
      if (event.key === 'ArrowRight') target = (index + 1) % scenarios.length;
      if (event.key === 'ArrowLeft') target = (index + scenarios.length - 1) % scenarios.length;
      if (event.key === 'Home') target = 0;
      if (event.key === 'End') target = scenarios.length - 1;
      if (target === undefined) return;
      event.preventDefault();
      chooseScenario(target);
      tabs.children[target].focus();
    });
    tabs.append(button);
    const option = document.createElement('option');
    option.value = scenario.id;
    option.textContent = scenario.label;
    promptKind.append(option);
  });

  function renderStep() {
    const scenario = current();
    const step = scenario.steps[state.step];
    setText('step-title', step.title);
    setText('step-context', step.context);
    setText('step-message', step.message);
    setText('step-action', step.action);
    setText('step-evidence', step.evidence);
    setText('step-skill', step.skill);
    setText('human-reply', step.reply || '');
    byId('human-turn').hidden = !step.reply;
    setText('step-count', `${scenario.steps.length} pasos en este ejemplo`);
    setText('step-position', `${state.step + 1} / ${scenario.steps.length}`);
    byId('step-prev').disabled = state.step === 0;
    byId('step-next').disabled = state.step === scenario.steps.length - 1;
    Array.from(journey.children).forEach((item, index) => {
      const button = item.firstElementChild;
      button.dataset.complete = String(index < state.step);
      if (index === state.step) button.setAttribute('aria-current', 'step');
      else button.removeAttribute('aria-current');
    });
  }

  function moveToStep(index) {
    state.step = index;
    renderStep();
    if (window.matchMedia('(max-width: 760px)').matches) {
      const stage = byId('step-stage');
      stage.tabIndex = -1;
      stage.focus({ preventScroll: true });
      stage.scrollIntoView({ block: 'start', behavior: 'auto' });
    }
  }

  function chooseScenario(index) {
    state.scenario = index;
    state.step = 0;
    const scenario = current();
    Array.from(tabs.children).forEach((tab, tabIndex) => {
      const selected = tabIndex === index;
      tab.setAttribute('aria-selected', String(selected));
      tab.tabIndex = selected ? 0 : -1;
    });
    byId('scenario-panel').setAttribute('aria-labelledby', `scenario-tab-${scenario.id}`);
    setText('scenario-kicker', scenario.kicker);
    setText('scenario-title', scenario.title);
    setText('scenario-summary', scenario.summary);
    setText('scenario-scope', scenario.scope);
    setText('scenario-prompt', scenario.prompt);
    setText('scenario-boundary', scenario.boundary);
    setText('final-message', scenario.final.message);
    setText('final-artifact', scenario.final.artifact);
    setText('final-acceptance', scenario.final.acceptance);
    setText('final-boundary', scenario.final.boundary);
    byId('scenario-result').open = false;
    journey.replaceChildren();
    scenario.steps.forEach((step, stepIndex) => {
      const item = document.createElement('li');
      const button = document.createElement('button');
      button.type = 'button';
      button.setAttribute('aria-label', `Paso ${stepIndex + 1}: ${step.title}`);
      const number = document.createElement('span');
      number.className = 'step-number';
      number.textContent = stepIndex + 1;
      number.setAttribute('aria-hidden', 'true');
      button.append(number, document.createTextNode(step.title));
      button.addEventListener('click', () => {
        moveToStep(stepIndex);
      });
      item.append(button);
      journey.append(item);
    });
    renderStep();
  }

  byId('step-prev').addEventListener('click', () => {
    if (state.step > 0) moveToStep(state.step - 1);
  });
  byId('step-next').addEventListener('click', () => {
    if (state.step < current().steps.length - 1) moveToStep(state.step + 1);
  });
  byId('show-result').addEventListener('click', () => {
    byId('scenario-result').open = true;
    byId('scenario-result').scrollIntoView({ block: 'nearest', behavior: 'auto' });
  });
  promptKind.addEventListener('change', () => {
    byId('prompt-text').value = scenarios.find((item) => item.id === promptKind.value).prompt;
    setText('copy-status', '');
  });
  byId('prompt-text').addEventListener('input', () => setText('copy-status', ''));
  byId('copy-prompt').addEventListener('click', async () => {
    const text = byId('prompt-text').value;
    if (!text.trim()) {
      setText('copy-status', 'Escribí un pedido antes de copiarlo.');
      byId('prompt-text').focus();
      return;
    }
    try {
      if (!navigator.clipboard?.writeText) throw new Error('Clipboard no disponible');
      await navigator.clipboard.writeText(text);
      setText('copy-status', 'Pedido copiado. Podés pegarlo en tu conversación de Codex.');
    } catch {
      byId('prompt-text').focus();
      byId('prompt-text').select();
      setText('copy-status', 'El navegador no permitió copiar automáticamente. El texto quedó seleccionado para copiarlo.');
    }
  });

  function renderDelegation() {
    const mode = byId('delegation-mode').value;
    byId('delegation-request').disabled = mode !== 'on_request';
    const requested = byId('delegation-request').checked;
    const available = byId('host-agents').checked;
    const forcedInline = byId('force-inline').checked;
    const permitted = !forcedInline && available && (mode === 'auto' || (mode === 'on_request' && requested));
    byId('parallel-branches').hidden = !permitted;
    byId('inline-branch').hidden = permitted;
    let explanation;
    if (forcedInline) explanation = 'La instrucción del usuario de trabajar inline prevalece. El principal mantiene ambas tareas, cualquiera sea el modo elegido.';
    else if (mode === 'off') explanation = 'Con off, todo queda en el principal. Puede investigar, implementar y comprobar sin transferirse roles ni crear subagentes.';
    else if (!available && mode === 'on_request' && requested) explanation = 'Se pidió delegación, pero faltan herramientas del host. Si esa delegación concreta es requisito del pedido, se informa la limitación; no se simula un agente.';
    else if (!available) explanation = 'La política permite considerar delegación, pero el host no ofrece las herramientas. El principal puede continuar inline dentro del alcance.';
    else if (mode === 'on_request' && !requested) explanation = 'on_request necesita un pedido explícito de delegación. La independencia de las tareas por sí sola no lo concede.';
    else if (mode === 'on_request') explanation = 'Con el pedido explícito y herramientas disponibles, podrían delegarse tareas independientes. El principal conserva contexto, integración y comprobación.';
    else explanation = 'auto permite considerar colaboración útil. Este ejemplo muestra dos módulos independientes; no obliga a delegar ni crea agentes automáticamente.';
    setText('delegation-explanation', explanation);
  }
  ['delegation-mode', 'delegation-request', 'host-agents', 'force-inline'].forEach((id) => {
    byId(id).addEventListener('change', renderDelegation);
  });

  if ('IntersectionObserver' in window) {
    const links = Array.from(document.querySelectorAll('.chapter-nav a'));
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return;
        links.forEach((link) => {
          if (link.hash === `#${entry.target.id}`) link.setAttribute('aria-current', 'location');
          else link.removeAttribute('aria-current');
        });
      });
    }, { rootMargin: '-15% 0px -65% 0px', threshold: 0 });
    links.forEach((link) => {
      const target = document.getElementById(link.hash.slice(1));
      if (target) observer.observe(target);
    });
    observer.observe(byId('inicio'));
  }
  byId('prompt-text').value = scenarios[0].prompt;
  chooseScenario(0);
  renderDelegation();
})();
