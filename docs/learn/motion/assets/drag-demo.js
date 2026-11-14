/*
 * Swipe-to-reply instrument. Shows a finger writing a shared value and a
 * spring settling it, which is the gesture pattern used in lesson 2.
 *
 * Markup contract:
 *   <div data-drag-demo data-open="-96" data-threshold="-64">
 *     <div class="reply-row" data-row>…</div>
 *     <div class="reply-flag" data-flag>reply</div>
 *     <span data-value></span>
 *   </div>
 */
(function () {
  var reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  function mount(root) {
    var row = root.querySelector("[data-row]");
    var flag = root.querySelector("[data-flag]");
    var valueEl = root.querySelector("[data-value]");
    var open = Number(root.dataset.open || -96);
    var threshold = Number(root.dataset.threshold || -64);

    var x = 0;
    var dragging = false;
    var startX = 0;
    var base = 0;
    var frame = 0;

    function paint() {
      row.style.transform = "translateX(" + x + "px)";
      if (valueEl) valueEl.textContent = Math.round(x) + "px";
      if (flag) {
        var progress = Math.min(1, Math.abs(x) / Math.abs(open));
        flag.style.opacity = String(0.25 + progress * 0.75);
        flag.style.transform = "scale(" + (0.85 + progress * 0.15) + ")";
      }
    }

    function settle(to) {
      if (frame) cancelAnimationFrame(frame);

      if (reduceMotion) {
        x = to;
        paint();
        return;
      }

      var v = 0;
      var last = performance.now();

      function step(now) {
        var elapsed = Math.min(0.05, (now - last) / 1000);
        last = now;
        var steps = Math.max(1, Math.round(elapsed / (1 / 240)));
        var dt = elapsed / steps;

        for (var i = 0; i < steps; i += 1) {
          var a = -420 * (x - to) - 34 * v;
          v += a * dt;
          x += v * dt;
        }

        paint();

        if (Math.abs(to - x) < 0.4 && Math.abs(v) < 4) {
          x = to;
          paint();
          frame = 0;
          return;
        }
        frame = requestAnimationFrame(step);
      }

      frame = requestAnimationFrame(step);
    }

    row.addEventListener("pointerdown", function (event) {
      if (frame) cancelAnimationFrame(frame);
      dragging = true;
      startX = event.clientX;
      base = x;
      row.setPointerCapture(event.pointerId);
    });

    row.addEventListener("pointermove", function (event) {
      if (!dragging) return;

      var next = base + (event.clientX - startX);

      /* Resistance past the open position, hard stop at rest. */
      if (next < open) next = open + (next - open) / 3;
      if (next > 0) next = next / 3;

      x = next;
      paint();
    });

    function end(event) {
      if (!dragging) return;
      dragging = false;
      if (row.hasPointerCapture && row.hasPointerCapture(event.pointerId)) {
        row.releasePointerCapture(event.pointerId);
      }
      settle(x <= threshold ? open : 0);
    }

    row.addEventListener("pointerup", end);
    row.addEventListener("pointercancel", end);

    root.querySelectorAll("[data-reset]").forEach(function (button) {
      button.addEventListener("click", function () {
        settle(0);
      });
    });

    paint();
  }

  document.querySelectorAll("[data-drag-demo]").forEach(mount);
})();
