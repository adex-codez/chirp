/*
 * Anchor instrument. A browser-side stand-in for a shared value plus one
 * animation function, so the lesson can be felt instead of imagined.
 *
 * It mirrors the two things Reanimated does: an animation function writes
 * a number over time, and a style reads that number every frame.
 *
 * Markup contract (see lessons for examples):
 *   <div data-anchor-demo data-value="1" data-unit="scale">
 *     <div class="stage-target" data-target></div>
 *     <button data-run="timing" data-to="0.94" data-duration="150">…</button>
 *     <span data-value></span><span data-derived></span>
 *   </div>
 */
(function () {
  var reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  var EASE_IN_OUT_QUAD = function (t) {
    return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
  };

  function format(n) {
    return (Math.round(n * 1000) / 1000).toFixed(3);
  }

  function mount(root) {
    var target = root.querySelector("[data-target]");
    var valueEl = root.querySelector("[data-value]");
    var derivedEl = root.querySelector("[data-derived]");
    var state = { value: Number(root.dataset.value || 1), frame: 0 };

    function paint() {
      if (target) {
        target.style.transform = "scale(" + state.value + ")";
      }
      if (valueEl) valueEl.textContent = format(state.value);
      if (derivedEl) {
        derivedEl.textContent =
          "transform: [{ " + (root.dataset.unit || "scale") + ": " + format(state.value) + " }]";
      }
    }

    function stop() {
      if (state.frame) cancelAnimationFrame(state.frame);
      state.frame = 0;
    }

    function runTiming(to, duration, done) {
      var from = state.value;
      var start = performance.now();

      function step(now) {
        var t = Math.min(1, (now - start) / duration);
        state.value = from + (to - from) * EASE_IN_OUT_QUAD(t);
        paint();
        if (t < 1) {
          state.frame = requestAnimationFrame(step);
        } else {
          state.value = to;
          paint();
          state.frame = 0;
          if (done) done();
        }
      }

      state.frame = requestAnimationFrame(step);
    }

    /* Fixed-step spring, integrated at 240 Hz then painted per frame. */
    function runSpring(to, config, done) {
      var stiffness = config.stiffness;
      var damping = config.damping;
      var mass = config.mass;
      var x = state.value;
      var v = config.velocity || 0;
      var last = performance.now();
      var accumulator = 0;
      var stepSize = 1 / 240;

      function step(now) {
        var elapsed = Math.min(0.064, (now - last) / 1000);
        last = now;
        accumulator += elapsed;

        while (accumulator >= stepSize) {
          var a = (-stiffness * (x - to) - damping * v) / mass;
          v += a * stepSize;
          x += v * stepSize;
          accumulator -= stepSize;
        }

        state.value = x;
        paint();

        if (Math.abs(to - x) < 0.0005 && Math.abs(v) < 0.005) {
          state.value = to;
          paint();
          state.frame = 0;
          if (done) done();
          return;
        }
        state.frame = requestAnimationFrame(step);
      }

      state.frame = requestAnimationFrame(step);
    }

    root.querySelectorAll("[data-run]").forEach(function (button) {
      button.addEventListener("click", function () {
        stop();
        var to = Number(button.dataset.to);
        var duration = Number(button.dataset.duration || 150);

        if (reduceMotion) {
          state.value = to;
          paint();
          return;
        }

        if (button.dataset.run === "spring") {
          runSpring(to, {
            stiffness: Number(button.dataset.stiffness || 900),
            damping: Number(button.dataset.damping || 120),
            mass: Number(button.dataset.mass || 4),
            velocity: Number(button.dataset.velocity || 0),
          });
        } else {
          runTiming(to, duration);
        }
      });
    });

    paint();
  }

  document.querySelectorAll("[data-anchor-demo]").forEach(mount);
})();
