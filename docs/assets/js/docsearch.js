import docsearch from "@docsearch/js";

docsearch({
  container: "#docsearch",
  appId: "BIQ7DDWR39",
  indexName: "Production",
  apiKey: "27590c872bf247526427720080358240",
  insights: true
});

const onClick = function () {
  document.getElementsByClassName("DocSearch-Button")[0].click();
};

document.getElementById("searchToggleMobile").onclick = onClick;
document.getElementById("searchToggleDesktop").onclick = onClick;
