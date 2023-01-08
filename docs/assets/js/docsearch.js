import docsearch from '@docsearch/js';

var searchPlaceholder = document.getElementById('search-placeholder');

if (searchPlaceholder !== null) {
  searchPlaceholder.className = 'd-none';
}

docsearch({
  container: '#docsearch',
  appId: 'OX8DV2T9J9',
  indexName: 'authelia_com',
  apiKey: 'a4d51f0dde6bb5a07a5bf6022408a353',
  debug: false,
});
