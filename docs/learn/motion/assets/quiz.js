/*
 * Quiz widget. Reusable across lessons.
 *
 * Markup contract:
 *   <section data-quiz>
 *     <span data-quiz-score></span>
 *     <div class="quiz-item">
 *       <p class="quiz-q">…</p>
 *       <ul class="quiz-options">
 *         <li><button data-correct="true" data-why="…">…</button></li>
 *       </ul>
 *       <p data-feedback aria-live="polite"></p>
 *     </div>
 *   </section>
 *
 * Options must be written to the same length so formatting leaks no clue.
 */
(function () {
  function mount(quiz) {
    var items = Array.prototype.slice.call(quiz.querySelectorAll(".quiz-item"));
    var scoreEl = quiz.querySelector("[data-quiz-score]");
    var answered = 0;
    var correctFirstTry = 0;

    items.forEach(function (item) {
      var options = Array.prototype.slice.call(item.querySelectorAll("button"));
      var feedback = item.querySelector("[data-feedback]");
      var done = false;
      var missed = false;

      options.forEach(function (option) {
        option.addEventListener("click", function () {
          if (done) return;

          var isCorrect = option.dataset.correct === "true";

          if (isCorrect) {
            option.dataset.state = "correct";
            done = true;
            options.forEach(function (other) {
              other.disabled = true;
            });
            feedback.innerHTML =
              "<b>Right.</b> " + (option.dataset.why || "");
            answered += 1;
            if (!missed) correctFirstTry += 1;
          } else {
            missed = true;
            option.dataset.state = "wrong";
            option.disabled = true;
            feedback.innerHTML =
              "<b>Not this one.</b> " + (option.dataset.why || "");
          }

          if (scoreEl) {
            scoreEl.textContent =
              correctFirstTry + " of " + items.length + " first try";
          }
        });
      });
    });

    if (scoreEl) scoreEl.textContent = "0 of " + items.length + " first try";
  }

  document.querySelectorAll("[data-quiz]").forEach(mount);
})();
