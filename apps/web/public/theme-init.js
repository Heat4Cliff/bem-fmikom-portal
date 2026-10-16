(() => {
  const allowed = new Set(["nawasena", "classic", "salt", "matcha", "mint"]);
  try {
    const stored =
      window.localStorage.getItem("bem-fmikom-theme") ?? window.localStorage.getItem("bem-theme");
    const theme = allowed.has(stored) ? stored : "nawasena";
    document.documentElement.dataset.theme = theme;
    if (theme !== "nawasena" && !window.localStorage.getItem("bem-fmikom-theme")) {
      window.localStorage.setItem("bem-fmikom-theme", theme);
    }
  } catch {
    document.documentElement.dataset.theme = "nawasena";
  }
})();
