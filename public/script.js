const translations = {
  en: document.querySelectorAll('[data-lang="en"]'),
  ru: document.querySelectorAll('[data-lang="ru"]'),
};

function setLanguage(lang) {
  document.documentElement.lang = lang;
  localStorage.setItem("lang", lang);

  const langToHide = lang === "en" ? "ru" : "en";

  translations[langToHide].forEach((el) => (el.style.display = "none"));
  translations[lang].forEach((el) => (el.style.display = ""));

  document.getElementById(`lang-${lang}`).classList.add("active");
  document.getElementById(`lang-${langToHide}`).classList.remove("active");

  if (window.event) event.preventDefault();
}

function setTheme(theme) {
  if (theme === "new") {
    document.documentElement.setAttribute("data-theme", "new");
  } else {
    document.documentElement.removeAttribute("data-theme");
  }

  localStorage.setItem("theme", theme);

  document
    .getElementById("theme-old")
    .classList.toggle("active", theme === "old");
  document
    .getElementById("theme-new")
    .classList.toggle("active", theme === "new");

  if (window.event) event.preventDefault();
}

function copyGPGKey() {
  var e = window.event;
  if (e) { e.preventDefault(); e.stopPropagation(); }
  const keyText = document.getElementById('gpg-key').textContent;
  navigator.clipboard.writeText(keyText).then(() => {
    const btnEn = document.getElementById('copy-gpg');
    const btnRu = document.getElementById('copy-gpg-ru');
    const prevEn = btnEn.textContent;
    const prevRu = btnRu.textContent;
    btnEn.textContent = 'Copied!';
    btnRu.textContent = 'Скопировано!';
    setTimeout(() => {
      btnEn.textContent = prevEn;
      btnRu.textContent = prevRu;
    }, 2000);
  });
}

document.addEventListener("DOMContentLoaded", () => {
  const savedLang = localStorage.getItem("lang") || "en";
  const savedTheme = localStorage.getItem("theme") || "old";

  setLanguage(savedLang);
  setTheme(savedTheme);
});
