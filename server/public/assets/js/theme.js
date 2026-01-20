// Theme toggle functionality
(function () {
  "use strict";

  // Get theme from localStorage or system preference
  function getInitialTheme() {
    const savedTheme = localStorage.getItem("theme");
    if (savedTheme) {
      return savedTheme;
    }

    // Check system preference
    if (
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches
    ) {
      return "dark";
    }

    return "light";
  }

  // Apply theme to document
  function applyTheme(theme) {
    if (theme === "dark") {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
    localStorage.setItem("theme", theme);
  }

  // Initialize theme before page renders (prevents flash)
  const initialTheme = getInitialTheme();
  applyTheme(initialTheme);

  // Toggle theme function (can be called from anywhere)
  window.toggleTheme = function () {
    const currentTheme = document.documentElement.classList.contains("dark")
      ? "dark"
      : "light";
    const newTheme = currentTheme === "dark" ? "light" : "dark";
    applyTheme(newTheme);

    // Dispatch event for other components to react
    window.dispatchEvent(
      new CustomEvent("themechange", { detail: { theme: newTheme } }),
    );
  };

  // Listen to system theme changes
  if (window.matchMedia) {
    window
      .matchMedia("(prefers-color-scheme: dark)")
      .addEventListener("change", (e) => {
        // Only auto-switch if user hasn't set a preference
        if (!localStorage.getItem("theme")) {
          applyTheme(e.matches ? "dark" : "light");
        }
      });
  }

  // Initialize theme toggle buttons when DOM is ready
  document.addEventListener("DOMContentLoaded", function () {
    const themeToggleBtns = document.querySelectorAll("[data-theme-toggle]");
    themeToggleBtns.forEach((btn) => {
      btn.addEventListener("click", window.toggleTheme);

      // Update icon based on current theme
      updateThemeIcon(btn);
    });

    // Listen for theme changes to update icons
    window.addEventListener("themechange", function () {
      themeToggleBtns.forEach(updateThemeIcon);
    });
  });

  function updateThemeIcon(btn) {
    const isDark = document.documentElement.classList.contains("dark");
    const sunIcon = btn.querySelector(".theme-icon-sun");
    const moonIcon = btn.querySelector(".theme-icon-moon");

    if (sunIcon && moonIcon) {
      if (isDark) {
        sunIcon.classList.remove("hidden");
        moonIcon.classList.add("hidden");
      } else {
        sunIcon.classList.add("hidden");
        moonIcon.classList.remove("hidden");
      }
    }
  }
})();
