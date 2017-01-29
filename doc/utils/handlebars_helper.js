define([
    'locales',
    'handlebars',
    'diffMatchPatch'
], function(locale, Handlebars, DiffMatchPatch) {

    /**
     * Return a text as markdown.
     * Currently only a little helper to replace apidoc-inline Links (#Group:Name).
     * Should be replaced with a full markdown lib.
     * @param string text
     */
    Handlebars.registerHelper('markdown', function(text) {
        if ( ! text ) {
          return text;
        }
        text = text.replace(/((\[(.*?)\])?\(#)((.+?):(.+?))(\))/mg, function(match, p1, p2, p3, p4, p5, p6) {
          var link = p3 || p5 + '/' + p6;
          return '<a href="#api-' + p5 + '-' + p6 + '">' + link + '</a>';
        });
        return text;
    });

    /**
     * start/stop timer for simple performance check.
     */
    var timer;
    Handlebars.registerHelper('startTimer', function(text) {
        timer = new Date();
        return '';
    });

    Handlebars.registerHelper('stopTimer', function(text) {
        console.log(new Date() - timer);
        return '';
    });

    /**
     * Return localized Text.
     * @param string text
     */
    Handlebars.registerHelper('__', function(text) {
        return locale.__(text);
    });

    /**
     * Console log.
     * @param mixed obj
     */
    Handlebars.registerHelper('cl', function(obj) {
        console.log(obj);
        return '';
    });

    /**
     * Replace underscore with space.
     * @param string text
     */
    Handlebars.registerHelper('underscoreToSpace', function(text) {
        return text.replace(/(_+)/g, ' ');
    });

    /**
     *
     */
    Handlebars.registerHelper('assign', function(name) {
        if(arguments.length > 0) {
            var type = typeof(arguments[1]);
            var arg = null;
            if(type === 'string' || type === 'number' || type === 'boolean') arg = arguments[1];
            Handlebars.registerHelper(name, function() { return arg; });
        }
        return '';
    });

    /**
     *
     */
    Handlebars.registerHelper('nl2br', function(text) {
        return _handlebarsNewlineToBreak(text);
    });

    /**
     *
     */
    Handlebars.registerHelper('if_eq', function(context, options) {
        var compare = context;
        // Get length if context is an object
        if (context instanceof Object && ! (options.hash.compare instanceof Object))
             compare = Object.keys(context).length;

        if (compare === options.hash.compare)
            return options.fn(this);

        return options.inverse(this);
    });

    /**
     *
     */
    Handlebars.registerHelper('if_gt', function(context, options) {
        var compare = context;
        // Get length if context is an object
        if (context instanceof Object && ! (options.hash.compare instanceof Object))
             compare = Object.keys(context).length;

        if(compare > options.hash.compare)
            return options.fn(this);

        return options.inverse(this);
    });

    /**
     *
     */
    var templateCache = {};
    Handlebars.registerHelper('subTemplate', function(name, sourceContext) {
        if ( ! templateCache[name])
            templateCache[name] = Handlebars.compile($('#template-' + name).html());

        var template = templateCache[name];
        var templateContext = $.extend({}, this, sourceContext.hash);
        return new Handlebars.SafeString( template(templateContext) );
    });

    /**
     *
     */
    Handlebars.registerHelper('toLowerCase', function(value) {
        return (value && typeof value === 'string') ? value.toLowerCase() : '';
    });

    /**
     *
     */
    Handlebars.registerHelper('splitFill', function(value, splitChar, fillChar) {
        var splits = value.split(splitChar);
        return new Array(splits.length).join(fillChar) + splits[splits.length - 1];
    });

    /**
     * Convert Newline to HTML-Break (nl2br).
     *
     * @param {String} text
     * @returns {String}
     */
    function _handlebarsNewlineToBreak(text) {
        return ('' + text).replace(/([^>\r\n]?)(\r\n|\n\r|\r|\n)/g, '$1' + '<br>' + '$2');
    }

    /**
     *
     */
    Handlebars.registerHelper('each_compare_list_field', function(source, compare, options) {
        var fieldName = options.hash.field;
        var newSource = [];
        if (source) {
            source.forEach(function(entry) {
                var values = entry;
                values['key'] = entry[fieldName];
                newSource.push(values);
            });
        }

        var newCompare = [];
        if (compare) {
            compare.forEach(function(entry) {
                var values = entry;
                values['key'] = entry[fieldName];
                newCompare.push(values);
            });
        }
        return _handlebarsEachCompared('key', newSource, newCompare, options);
    });

    /**
     *
     */
    Handlebars.registerHelper('each_compare_keys', function(source, compare, options) {
        var newSource = [];
        if (source) {
            var sourceFields = Object.keys(source);
            sourceFields.forEach(function(name) {
                var values = {};
                values['value'] = source[name];
                values['key'] = name;
                newSource.push(values);
            });
        }

        var newCompare = [];
        if (compare) {
            var compareFields = Object.keys(compare);
            compareFields.forEach(function(name) {
                var values = {};
                values['value'] = compare[name];
                values['key'] = name;
                newCompare.push(values);
            });
        }
        return _handlebarsEachCompared('key', newSource, newCompare, options);
    });

    /**
     *
     */
    Handlebars.registerHelper('each_compare_field', function(source, compare, options) {
        return _handlebarsEachCompared('field', source, compare, options);
    });

    /**
     *
     */
    Handlebars.registerHelper('each_compare_title', function(source, compare, options) {
        return _handlebarsEachCompared('title', source, compare, options);
    });

    /**
     *
     */
    Handlebars.registerHelper('reformat', function(source, type){
        if (type == 'json')
            try {
               return JSON.stringify(JSON.parse(source.trim()),null, "    ");
            } catch(e) {

            }
        return source
    });

    /**
     *
     */
    Handlebars.registerHelper('showDiff', function(source, compare, options) {
        var ds = '';
        if(source === compare) {
            ds = source;
        } else {
            if( ! source)
                return compare;

            if( ! compare)
                return source;

            var d = diffMatchPatch.diff_main(compare, source);
            diffMatchPatch.diff_cleanupSemantic(d);
            ds = diffMatchPatch.diff_prettyHtml(d);
            ds = ds.replace(/&para;/gm, '');
        }
        if(options === 'nl2br')
            ds = _handlebarsNewlineToBreak(ds);

        return ds;
    });

    /**
     *
     */
    function _handlebarsEachCompared(fieldname, source, compare, options)
    {
        var dataList = [];
        var index = 0;
        if(source) {
            source.forEach(function(sourceEntry) {
                var found = false;
                if (compare) {
                    compare.forEach(function(compareEntry) {
                        if(sourceEntry[fieldname] === compareEntry[fieldname]) {
                            var data = {
                                typeSame: true,
                                source: sourceEntry,
                                compare: compareEntry,
                                index: index
                            };
                            dataList.push(data);
                            found = true;
                            index++;
                        }
                    });
                }
                if ( ! found) {
                    var data = {
                        typeIns: true,
                        source: sourceEntry,
                        index: index
                    };
                    dataList.push(data);
                    index++;
                }
            });
        }

        if (compare) {
            compare.forEach(function(compareEntry) {
                var found = false;
                if (source) {
                    source.forEach(function(sourceEntry) {
                        if(sourceEntry[fieldname] === compareEntry[fieldname])
                            found = true;
                    });
                }
                if ( ! found) {
                    var data = {
                        typeDel: true,
                        compare: compareEntry,
                        index: index
                    };
                    dataList.push(data);
                    index++;
                }
            });
        }

        var ret = '';
        var length = dataList.length;
        for (var index in dataList) {
            if(index == (length - 1))
                dataList[index]['_last'] = true;
            ret = ret + options.fn(dataList[index]);
        }
        return ret;
    }

    var diffMatchPatch = new DiffMatchPatch();

    /**
     * Overwrite Colors
     */
    DiffMatchPatch.prototype.diff_prettyHtml = function(diffs) {
      var html = [];
      var pattern_amp = /&/g;
      var pattern_lt = /</g;
      var pattern_gt = />/g;
      var pattern_para = /\n/g;
      for (var x = 0; x < diffs.length; x++) {
        var op = diffs[x][0];    // Operation (insert, delete, equal)
        var data = diffs[x][1];  // Text of change.
        var text = data.replace(pattern_amp, '&amp;').replace(pattern_lt, '&lt;')
            .replace(pattern_gt, '&gt;').replace(pattern_para, '&para;<br>');
        switch (op) {
          case DIFF_INSERT:
            html[x] = '<ins>' + text + '</ins>';
            break;
          case DIFF_DELETE:
            html[x] = '<del>' + text + '</del>';
            break;
          case DIFF_EQUAL:
            html[x] = '<span>' + text + '</span>';
            break;
        }
      }
      return html.join('');
    };

    // Exports
    return Handlebars;
});
