// SPDX-License-Identifier: AGPL-3.0-or-later
// import-accordion.js — one-card-open-at-a-time accordion for the left
// column of import.php. Each `.accordion-card` group is named by the
// parent `[data-accordion]` value; clicking a header toggles that card and
// collapses the others in the same group. Page load honours the initial
// `aria-expanded` value the dispatcher writes into the markup.

(function () {
  function init() {
    // Group cards by their parent's data-accordion attribute.
    var groups = {};
    document.querySelectorAll('.accordion-card').forEach(function (card) {
      var group = (card.closest('[data-accordion]') || {}).getAttribute && card.closest('[data-accordion]').getAttribute('data-accordion');
      if (!group) return;
      (groups[group] = groups[group] || []).push(card);
    });

    Object.keys(groups).forEach(function (groupName) {
      var cards = groups[groupName];

      // Apply initial open state from aria-expanded on the toggle button.
      cards.forEach(function (card) {
        var toggle = card.querySelector('.accordion-toggle');
        var open = toggle && toggle.getAttribute('aria-expanded') === 'true';
        card.classList.toggle('is-open', open);
      });

      // Wire click: opening one card collapses all other cards in the group.
      cards.forEach(function (card) {
        var toggle = card.querySelector('.accordion-toggle');
        if (!toggle) return;
        toggle.addEventListener('click', function () {
          var wasOpen = card.classList.contains('is-open');
          cards.forEach(function (c) {
            var t = c.querySelector('.accordion-toggle');
            c.classList.remove('is-open');
            if (t) t.setAttribute('aria-expanded', 'false');
          });
          if (!wasOpen) {
            card.classList.add('is-open');
            toggle.setAttribute('aria-expanded', 'true');
          }
        });
      });
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();