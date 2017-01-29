require.config({
    paths: {
        bootstrap: './vendor/bootstrap.min',
        diffMatchPatch: './vendor/diff_match_patch.min',
        handlebars: './vendor/handlebars.min',
        handlebarsExtended: './utils/handlebars_helper',
        jquery: './vendor/jquery.min',
        locales: './locales/locale',
        lodash: './vendor/lodash.custom.min',
        pathToRegexp: './vendor/path-to-regexp/index',
        prettify: './vendor/prettify/prettify',
        semver: './vendor/semver.min',
        utilsSampleRequest: './utils/send_sample_request',
        webfontloader: './vendor/webfontloader',
        list: './vendor/list.min'
    },
    shim: {
        bootstrap: {
            deps: ['jquery']
        },
        diffMatchPatch: {
            exports: 'diff_match_patch'
        },
        handlebars: {
            exports: 'Handlebars'
        },
        handlebarsExtended: {
            deps: ['jquery', 'handlebars'],
            exports: 'Handlebars'
        },
        prettify: {
            exports: 'prettyPrint'
        }
    },
    urlArgs: 'v=' + (new Date()).getTime(),
    waitSeconds: 15
});

require([
    'jquery',
    'lodash',
    'locales',
    'handlebarsExtended',
    './api_project.js',
    './api_data.js',
    'prettify',
    'utilsSampleRequest',
    'semver',
    'webfontloader',
    'bootstrap',
    'pathToRegexp',
    'list'
], function($, _, locale, Handlebars, apiProject, apiData, prettyPrint, sampleRequest, semver, WebFont) {

    // load google web fonts
    loadGoogleFontCss();

    var api = apiData.api;

    //
    // Templates
    //
    var templateHeader         = Handlebars.compile( $('#template-header').html() );
    var templateFooter         = Handlebars.compile( $('#template-footer').html() );
    var templateArticle        = Handlebars.compile( $('#template-article').html() );
    var templateCompareArticle = Handlebars.compile( $('#template-compare-article').html() );
    var templateGenerator      = Handlebars.compile( $('#template-generator').html() );
    var templateProject        = Handlebars.compile( $('#template-project').html() );
    var templateSections       = Handlebars.compile( $('#template-sections').html() );
    var templateSidenav        = Handlebars.compile( $('#template-sidenav').html() );

    //
    // apiProject defaults
    //
    if ( ! apiProject.template)
        apiProject.template = {};

    if (apiProject.template.withCompare == null)
        apiProject.template.withCompare = true;

    if (apiProject.template.withGenerator == null)
        apiProject.template.withGenerator = true;

    if (apiProject.template.forceLanguage)
        locale.setLanguage(apiProject.template.forceLanguage);

    // Setup jQuery Ajax
    $.ajaxSetup(apiProject.template.jQueryAjaxSetup);

    //
    // Data transform
    //
    // grouped by group
    var apiByGroup = _.groupBy(api, function(entry) {
        return entry.group;
    });

    // grouped by group and name
    var apiByGroupAndName = {};
    $.each(apiByGroup, function(index, entries) {
        apiByGroupAndName[index] = _.groupBy(entries, function(entry) {
            return entry.name;
        });
    });

    //
    // sort api within a group by title ASC and custom order
    //
    var newList = [];
    var umlauts = { 'ä': 'ae', 'ü': 'ue', 'ö': 'oe', 'ß': 'ss' }; // TODO: remove in version 1.0
    $.each (apiByGroupAndName, function(index, groupEntries) {
        // get titles from the first entry of group[].name[] (name has versioning)
        var titles = [];
        $.each (groupEntries, function(titleName, entries) {
            var title = entries[0].title;
            if(title !== undefined) {
                title.toLowerCase().replace(/[äöüß]/g, function($0) { return umlauts[$0]; });
                titles.push(title + '#~#' + titleName); // '#~#' keep reference to titleName after sorting
            }
        });
        // sort by name ASC
        titles.sort();

        // custom order
        if (apiProject.order)
            titles = sortByOrder(titles, apiProject.order, '#~#');

        // add single elements to the new list
        titles.forEach(function(name) {
            var values = name.split('#~#');
            var key = values[1];
            groupEntries[key].forEach(function(entry) {
                newList.push(entry);
            });
        });
    });
    // api overwrite with ordered list
    api = newList;

    //
    // Group- and Versionlists
    //
    var apiGroups = {};
    var apiGroupTitles = {};
    var apiVersions = {};
    apiVersions[apiProject.version] = 1;

    $.each(api, function(index, entry) {
        apiGroups[entry.group] = 1;
        apiGroupTitles[entry.group] = entry.groupTitle || entry.group;
        apiVersions[entry.version] = 1;
    });

    // sort groups
    apiGroups = Object.keys(apiGroups);
    apiGroups.sort();

    // custom order
    if (apiProject.order)
        apiGroups = sortByOrder(apiGroups, apiProject.order);

    // sort versions DESC
    apiVersions = Object.keys(apiVersions);
    apiVersions.sort(semver.compare);
    apiVersions.reverse();

    //
    // create Navigationlist
    //
    var nav = [];
    apiGroups.forEach(function(group) {
        // Mainmenu entry
        nav.push({
            group: group,
            isHeader: true,
            title: apiGroupTitles[group]
        });

        // Submenu
        var oldName = '';
        api.forEach(function(entry) {
            if (entry.group === group) {
                if (oldName !== entry.name) {
                    nav.push({
                        title: entry.title,
                        group: group,
                        name: entry.name,
                        type: entry.type,
                        version: entry.version
                    });
                } else {
                    nav.push({
                        title: entry.title,
                        group: group,
                        hidden: true,
                        name: entry.name,
                        type: entry.type,
                        version: entry.version
                    });
                }
                oldName = entry.name;
            }
        });
    });

    /**
     * Add navigation items by analyzing the HTML content and searching for h1 and h2 tags
     * @param nav Object the navigation array
     * @param content string the compiled HTML content
     * @param index where to insert items
     * @return boolean true if any good-looking (i.e. with a group identifier) <h1> tag was found
     */
    function add_nav(nav, content, index) {
        var found_level1 = false;
        if ( ! content) {
          return found_level1;
        }
        var topics = content.match(/<h(1|2).*?>(.+?)<\/h(1|2)>/gi);
        if ( topics ) {
          topics.forEach(function(entry) {
              var level = entry.substring(2,3);
              var title = entry.replace(/<.+?>/g, '');    // Remove all HTML tags for the title
              var entry_tags = entry.match(/id="api-([^\-]+)(?:-(.+))?"/);    // Find the group and name in the id property
              var group = (entry_tags ? entry_tags[1] : null);
              var name = (entry_tags ? entry_tags[2] : null);
              if (level==1 && title && group)  {
                  nav.splice(index, 0, {
                      group: group,
                      isHeader: true,
                      title: title,
                      isFixed: true
                  });
                  index++;
                  found_level1 = true;
              }
              if (level==2 && title && group && name)    {
                  nav.splice(index, 0, {
                      group: group,
                      name: name,
                      isHeader: false,
                      title: title,
                      isFixed: false,
                      version: '1.0'
                  });
                  index++;
              }
          });
        }
        return found_level1;
    }

    // Mainmenu Header entry
    if (apiProject.header) {
        var found_level1 = add_nav(nav, apiProject.header.content, 0); // Add level 1 and 2 titles
        if (!found_level1) {    // If no Level 1 tags were found, make a title
            nav.unshift({
                group: '_',
                isHeader: true,
                title: (apiProject.header.title == null) ? locale.__('General') : apiProject.header.title,
                isFixed: true
            });
        }
    }

    // Mainmenu Footer entry
    if (apiProject.footer) {
        var last_nav_index = nav.length;
        var found_level1 = add_nav(nav, apiProject.footer.content, nav.length); // Add level 1 and 2 titles
        if (!found_level1 && apiProject.footer.title != null) {    // If no Level 1 tags were found, make a title
            nav.splice(last_nav_index, 0, {
                group: '_footer',
                isHeader: true,
                title: apiProject.footer.title,
                isFixed: true
            });
        }
    }

    // render pagetitle
    var title = apiProject.title ? apiProject.title : 'apiDoc: ' + apiProject.name + ' - ' + apiProject.version;
    $(document).attr('title', title);

    // remove loader
    $('#loader').remove();

    // render sidenav
    var fields = {
        nav: nav
    };
    $('#sidenav').append( templateSidenav(fields) );

    // render Generator
    $('#generator').append( templateGenerator(apiProject) );

    // render Project
    _.extend(apiProject, { versions: apiVersions});
    $('#project').append( templateProject(apiProject) );

    // render apiDoc, header/footer documentation
    if (apiProject.header)
        $('#header').append( templateHeader(apiProject.header) );

    if (apiProject.footer)
        $('#footer').append( templateFooter(apiProject.footer) );

    //
    // Render Sections and Articles
    //
    var articleVersions = {};
    var content = '';
    apiGroups.forEach(function(groupEntry) {
        var articles = [];
        var oldName = '';
        var fields = {};
        var title = groupEntry;
        var description = '';
        articleVersions[groupEntry] = {};

        // render all articles of a group
        api.forEach(function(entry) {
            if(groupEntry === entry.group) {
                if (oldName !== entry.name) {
                    // determine versions
                    api.forEach(function(versionEntry) {
                        if (groupEntry === versionEntry.group && entry.name === versionEntry.name) {
                            if ( ! articleVersions[entry.group].hasOwnProperty(entry.name) ) {
                                articleVersions[entry.group][entry.name] = [];
                            }
                            articleVersions[entry.group][entry.name].push(versionEntry.version);
                        }
                    });
                    fields = {
                        article: entry,
                        versions: articleVersions[entry.group][entry.name]
                    };
                } else {
                    fields = {
                        article: entry,
                        hidden: true,
                        versions: articleVersions[entry.group][entry.name]
                    };
                }

                // add prefix URL for endpoint
                if (apiProject.url)
                    fields.article.url = apiProject.url + fields.article.url;

                addArticleSettings(fields, entry);

                if (entry.groupTitle)
                    title = entry.groupTitle;

                // TODO: make groupDescription compareable with older versions (not important for the moment)
                if (entry.groupDescription)
                    description = entry.groupDescription;

                articles.push({
                    article: templateArticle(fields),
                    group: entry.group,
                    name: entry.name
                });
                oldName = entry.name;
            }
        });

        // render Section with Articles
        var fields = {
            group: groupEntry,
            title: title,
            description: description,
            articles: articles
        };
        content += templateSections(fields);
    });
    $('#sections').append( content );

    // Bootstrap Scrollspy
    $(this).scrollspy({ target: '#scrollingNav', offset: 18 });

    // Content-Scroll on Navigation click.
    $('.sidenav').find('a').on('click', function(e) {
        e.preventDefault();
        var id = $(this).attr('href');
        if ($(id).length > 0)
            $('html,body').animate({ scrollTop: parseInt($(id).offset().top) }, 400);
        window.location.hash = $(this).attr('href');
    });

    // Quickjump on Pageload to hash position.
    if(window.location.hash) {
        var id = window.location.hash;
        if ($(id).length > 0)
            $('html,body').animate({ scrollTop: parseInt($(id).offset().top) }, 0);
    }

    /**
     * Check if Parameter (sub) List has a type Field.
     * Example: @apiSuccess          varname1 No type.
     *          @apiSuccess {String} varname2 With type.
     *
     * @param {Object} fields
     */
    function _hasTypeInFields(fields) {
        var result = false;
        $.each(fields, function(name) {
            result = result || _.some(fields[name], function(item) { return item.type; });
        });
        return result;
    }

    /**
     * On Template changes, recall plugins.
     */
    function initDynamic() {
        // Bootstrap popover
        $('button[data-toggle="popover"]').popover().click(function(e) {
            e.preventDefault();
        });

        var version = $('#version strong').html();
        $('#sidenav li').removeClass('is-new');
        if (apiProject.template.withCompare) {
            $('#sidenav li[data-version=\'' + version + '\']').each(function(){
                var group = $(this).data('group');
                var name = $(this).data('name');
                var length = $('#sidenav li[data-group=\'' + group + '\'][data-name=\'' + name + '\']').length;
                var index  = $('#sidenav li[data-group=\'' + group + '\'][data-name=\'' + name + '\']').index($(this));
                if (length === 1 || index === (length - 1))
                    $(this).addClass('is-new');
            });
        }

        // tabs
        $('.nav-tabs-examples a').click(function (e) {
            e.preventDefault();
            $(this).tab('show');
        });
        $('.nav-tabs-examples').find('a:first').tab('show');

        // sample request switch
        $('.sample-request-switch').click(function (e) {
            var name = '.' + $(this).attr('name') + '-fields';
            $(name).addClass('hide');
            $(this).parent().next(name).removeClass('hide');
        });

        // call scrollspy refresh method
        $(window).scrollspy('refresh');

        // init modules
        sampleRequest.initDynamic();
    }
    initDynamic();

    // Pre- / Code-Format
    prettyPrint();

    //
    // HTML-Template specific jQuery-Functions
    //
    // Change Main Version
    $('#versions li.version a').on('click', function(e) {
        e.preventDefault();

        var selectedVersion = $(this).html();
        $('#version strong').html(selectedVersion);

        // hide all
        $('article').addClass('hide');
        $('#sidenav li:not(.nav-fixed)').addClass('hide');

        // show 1st equal or lower Version of each entry
        $('article[data-version]').each(function(index) {
            var group = $(this).data('group');
            var name = $(this).data('name');
            var version = $(this).data('version');

            if (semver.lte(version, selectedVersion)) {
                if ($('article[data-group=\'' + group + '\'][data-name=\'' + name + '\']:visible').length === 0) {
                    // enable Article
                    $('article[data-group=\'' + group + '\'][data-name=\'' + name + '\'][data-version=\'' + version + '\']').removeClass('hide');
                    // enable Navigation
                    $('#sidenav li[data-group=\'' + group + '\'][data-name=\'' + name + '\'][data-version=\'' + version + '\']').removeClass('hide');
                    $('#sidenav li.nav-header[data-group=\'' + group + '\']').removeClass('hide');
                }
            }
        });

        // show 1st equal or lower Version of each entry
        $('article[data-version]').each(function(index) {
            var group = $(this).data('group');
            $('section#api-' + group).removeClass('hide');
            if ($('section#api-' + group + ' article:visible').length === 0) {
                $('section#api-' + group).addClass('hide');
            } else {
                $('section#api-' + group).removeClass('hide');
            }
        });

        initDynamic();
        return;
    });

    // compare all article with their predecessor
    $('#compareAllWithPredecessor').on('click', changeAllVersionCompareTo);

    // change version of an article
    $('article .versions li.version a').on('click', changeVersionCompareTo);

    // compare url-parameter
    $.urlParam = function(name) {
        var results = new RegExp('[\\?&amp;]' + name + '=([^&amp;#]*)').exec(window.location.href);
        return (results && results[1]) ? results[1] : null;
    };

    if ($.urlParam('compare')) {
        // URL Paramter ?compare=1 is set
        $('#compareAllWithPredecessor').trigger('click');

        if (window.location.hash) {
            var id = window.location.hash;
            $('html,body').animate({ scrollTop: parseInt($(id).offset().top) - 18 }, 0);
        }
    }

    /**
     * Initialize search
     */
    var options = {
      valueNames: [ 'nav-list-item' ]
    };
    var endpointsList = new List('scrollingNav', options);

    /**
     * Set initial focus to search input
     */
    $('#scrollingNav .sidenav-search input.search').focus();

    /**
     * Detect ESC key to reset search
     */
    $(document).keyup(function(e) {
      if (e.keyCode === 27) $('span.search-reset').click();
    });

    /**
     * Search reset
     */
    $('span.search-reset').on('click', function() {
      $('#scrollingNav .sidenav-search input.search')
        .val("")
        .focus()
      ;
      endpointsList.search();
    });

    /**
     * Change version of an article to compare it to an other version.
     */
    function changeVersionCompareTo(e) {
        e.preventDefault();

        var $root = $(this).parents('article');
        var selectedVersion = $(this).html();
        var $button = $root.find('.version');
        var currentVersion = $button.find('strong').html();
        $button.find('strong').html(selectedVersion);

        var group = $root.data('group');
        var name = $root.data('name');
        var version = $root.data('version');

        var compareVersion = $root.data('compare-version');

        if (compareVersion === selectedVersion)
            return;

        if ( ! compareVersion && version == selectedVersion)
            return;

        if (compareVersion && articleVersions[group][name][0] === selectedVersion || version === selectedVersion) {
            // the version of the entry is set to the highest version (reset)
            resetArticle(group, name, version);
        } else {
            var $compareToArticle = $('article[data-group=\'' + group + '\'][data-name=\'' + name + '\'][data-version=\'' + selectedVersion + '\']');

            var sourceEntry = {};
            var compareEntry = {};
            $.each(apiByGroupAndName[group][name], function(index, entry) {
                if (entry.version === version)
                    sourceEntry = entry;
                if (entry.version === selectedVersion)
                    compareEntry = entry;
            });

            var fields = {
                article: sourceEntry,
                compare: compareEntry,
                versions: articleVersions[group][name]
            };

            // add unique id
            // TODO: replace all group-name-version in template with id.
            fields.article.id = fields.article.group + '-' + fields.article.name + '-' + fields.article.version;
            fields.article.id = fields.article.id.replace(/\./g, '_');

            fields.compare.id = fields.compare.group + '-' + fields.compare.name + '-' + fields.compare.version;
            fields.compare.id = fields.compare.id.replace(/\./g, '_');

            var entry = sourceEntry;
            if (entry.parameter && entry.parameter.fields)
                fields._hasTypeInParameterFields = _hasTypeInFields(entry.parameter.fields);

            if (entry.error && entry.error.fields)
                fields._hasTypeInErrorFields = _hasTypeInFields(entry.error.fields);

            if (entry.success && entry.success.fields)
                fields._hasTypeInSuccessFields = _hasTypeInFields(entry.success.fields);

            if (entry.info && entry.info.fields)
                fields._hasTypeInInfoFields = _hasTypeInFields(entry.info.fields);

            var entry = compareEntry;
            if (fields._hasTypeInParameterFields !== true && entry.parameter && entry.parameter.fields)
                fields._hasTypeInParameterFields = _hasTypeInFields(entry.parameter.fields);

            if (fields._hasTypeInErrorFields !== true && entry.error && entry.error.fields)
                fields._hasTypeInErrorFields = _hasTypeInFields(entry.error.fields);

            if (fields._hasTypeInSuccessFields !== true && entry.success && entry.success.fields)
                fields._hasTypeInSuccessFields = _hasTypeInFields(entry.success.fields);

            if (fields._hasTypeInInfoFields !== true && entry.info && entry.info.fields)
                fields._hasTypeInInfoFields = _hasTypeInFields(entry.info.fields);

            var content = templateCompareArticle(fields);
            $root.after(content);
            var $content = $root.next();

            // Event on.click re-assign
            $content.find('.versions li.version a').on('click', changeVersionCompareTo);

            // select navigation
            $('#sidenav li[data-group=\'' + group + '\'][data-name=\'' + name + '\'][data-version=\'' + currentVersion + '\']').addClass('has-modifications');

            $root.remove();
            // TODO: on change main version or select the highest version re-render
        }

        initDynamic();
    }

    /**
     * Compare all currently selected Versions with their predecessor.
     */
    function changeAllVersionCompareTo(e) {
        e.preventDefault();
        $('article:visible .versions').each(function(){
            var $root = $(this).parents('article');
            var currentVersion = $root.data('version');
            var $foundElement = null;
            $(this).find('li.version a').each(function() {
                var selectVersion = $(this).html();
                if (selectVersion < currentVersion && ! $foundElement)
                    $foundElement = $(this);
            });

            if($foundElement)
                $foundElement.trigger('click');
        });
        initDynamic();
    }

    /**
     * Sort the fields.
     */
    function sortFields(fields_object) {
        $.each(fields_object, function (key, fields) {

            var reversed = fields.slice().reverse()

            var max_dot_count = Math.max.apply(null, reversed.map(function (item) {
                return item.field.split(".").length - 1;
            }))

            for (var dot_count = 1; dot_count <= max_dot_count; dot_count++) {
                reversed.forEach(function (item, index) {
                    var parts = item.field.split(".");
                    if (parts.length - 1 == dot_count) {
                        var fields_names = fields.map(function (item) { return item.field; });
                        if (parts.slice(1).length  >= 1) {
                            var prefix = parts.slice(0, parts.length - 1).join(".");
                            var prefix_index = fields_names.indexOf(prefix);
                            if (prefix_index > -1) {
                                fields.splice(fields_names.indexOf(item.field), 1);
                                fields.splice(prefix_index + 1, 0, item);
                            }
                        }
                    }
                });
            }
        });
    }

    /**
     * Add article settings.
     */
    function addArticleSettings(fields, entry) {
        // add unique id
        // TODO: replace all group-name-version in template with id.
        fields.id = fields.article.group + '-' + fields.article.name + '-' + fields.article.version;
        fields.id = fields.id.replace(/\./g, '_');

        if (entry.header && entry.header.fields) {
            sortFields(entry.header.fields);
            fields._hasTypeInHeaderFields = _hasTypeInFields(entry.header.fields);
        }

        if (entry.parameter && entry.parameter.fields) {
            sortFields(entry.parameter.fields);
            fields._hasTypeInParameterFields = _hasTypeInFields(entry.parameter.fields);
        }

        if (entry.error && entry.error.fields) {
            sortFields(entry.error.fields);
            fields._hasTypeInErrorFields = _hasTypeInFields(entry.error.fields);
        }

        if (entry.success && entry.success.fields) {
            sortFields(entry.success.fields);
            fields._hasTypeInSuccessFields = _hasTypeInFields(entry.success.fields);
        }

        if (entry.info && entry.info.fields) {
            sortFields(entry.info.fields);
            fields._hasTypeInInfoFields = _hasTypeInFields(entry.info.fields);
        }

        // add template settings
        fields.template = apiProject.template;
    }

    /**
     * Render Article.
     */
    function renderArticle(group, name, version) {
        var entry = {};
        $.each(apiByGroupAndName[group][name], function(index, currentEntry) {
            if (currentEntry.version === version)
                entry = currentEntry;
        });
        var fields = {
            article: entry,
            versions: articleVersions[group][name]
        };

        addArticleSettings(fields, entry);

        return templateArticle(fields);
    }

    /**
     * Render original Article and remove the current visible Article.
     */
    function resetArticle(group, name, version) {
        var $root = $('article[data-group=\'' + group + '\'][data-name=\'' + name + '\']:visible');
        var content = renderArticle(group, name, version);

        $root.after(content);
        var $content = $root.next();

        // Event on.click muss neu zugewiesen werden (sollte eigentlich mit on automatisch funktionieren... sollte)
        $content.find('.versions li.version a').on('click', changeVersionCompareTo);

        $('#sidenav li[data-group=\'' + group + '\'][data-name=\'' + name + '\'][data-version=\'' + version + '\']').removeClass('has-modifications');

        $root.remove();
        return;
    }

    /**
     * Load google fonts.
     */
    function loadGoogleFontCss() {
        WebFont.load({
            active: function() {
                // Update scrollspy
                $(window).scrollspy('refresh')
            },
            google: {
                families: ['Source Code Pro', 'Source Sans Pro:n4,n6,n7']
            }
        });
    }

    /**
     * Return ordered entries by custom order and append not defined entries to the end.
     * @param  {String[]} elements
     * @param  {String[]} order
     * @param  {String}   splitBy
     * @return {String[]} Custom ordered list.
     */
    function sortByOrder(elements, order, splitBy) {
        var results = [];
        order.forEach (function(name) {
            if (splitBy)
                elements.forEach (function(element) {
                    var parts = element.split(splitBy);
                    var key = parts[1]; // reference keep for sorting
                    if (key == name)
                        results.push(element);
                });
            else
                elements.forEach (function(key) {
                    if (key == name)
                        results.push(name);
                });
        });
        // Append all other entries that ar not defined in order
        elements.forEach(function(element) {
            if (results.indexOf(element) === -1)
                results.push(element);
        });
        return results;
    }

});
