define([
    './locales/ca.js',
    './locales/de.js',
    './locales/es.js',
    './locales/fr.js',
    './locales/it.js',
    './locales/nl.js',
    './locales/pl.js',
    './locales/pt_br.js',
    './locales/ro.js',
    './locales/ru.js',
    './locales/tr.js',
    './locales/vi.js',
    './locales/zh.js',
    './locales/zh_cn.js'
], function() {
    var langId = (navigator.language || navigator.userLanguage).toLowerCase().replace('-', '_');
    var language = langId.substr(0, 2);
    var locales = {};

    for (index in arguments) {
        for (property in arguments[index])
            locales[property] = arguments[index][property];
    }
    if ( ! locales['en'])
        locales['en'] = {};

    if ( ! locales[langId] && ! locales[language])
        language = 'en';

    var locale = (locales[langId] ? locales[langId] : locales[language]);

    function __(text) {
        var index = locale[text];
        if (index === undefined)
            return text;
        return index;
    };

    function setLanguage(language) {
        locale = locales[language];
    }

    return {
        __         : __,
        locales    : locales,
        locale     : locale,
        setLanguage: setLanguage
    };
});
