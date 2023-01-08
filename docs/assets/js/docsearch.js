import docsearch from '@docsearch/js';

var searchPlaceholder = document.getElementById('search-placeholder');

if (searchPlaceholder !== null) {
  searchPlaceholder.className = 'd-none';
}

docsearch({
  container: '#docsearch',
  appId: '{{ os.Getenv "ALGOLIA_APP_ID" }}',
  indexName: '{{ os.Getenv "ALGOLIA_INDEX_NAME" }}',
  apiKey: '{{ os.Getenv "ALGOLIA_API_KEY" }}',
  debug: false,
});
