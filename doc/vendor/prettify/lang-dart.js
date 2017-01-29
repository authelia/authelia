/*

 Copyright (C) 2013 Google Inc.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/
PR.registerLangHandler(PR.createSimpleLexer([["pln",/^[\t\n\r \xA0]+/,null,"\t\n\r \u00a0"]],[["com",/^#!(?:.*)/],["kwd",/^\b(?:import|library|part of|part|as|show|hide)\b/i],["com",/^\/\/(?:.*)/],["com",/^\/\*[^*]*\*+(?:[^\/*][^*]*\*+)*\//],["kwd",/^\b(?:class|interface)\b/i],["kwd",/^\b(?:assert|async|await|break|case|catch|continue|default|do|else|finally|for|if|in|is|new|return|super|switch|sync|this|throw|try|while)\b/i],["kwd",/^\b(?:abstract|const|extends|factory|final|get|implements|native|operator|set|static|typedef|var)\b/i],
["typ",/^\b(?:bool|double|Dynamic|int|num|Object|String|void)\b/i],["kwd",/^\b(?:false|null|true)\b/i],["str",/^r?[\']{3}[\s|\S]*?[^\\][\']{3}/],["str",/^r?[\"]{3}[\s|\S]*?[^\\][\"]{3}/],["str",/^r?\'(\'|(?:[^\n\r\f])*?[^\\]\')/],["str",/^r?\"(\"|(?:[^\n\r\f])*?[^\\]\")/],["typ",/^[A-Z]\w*/],["pln",/^[a-z_$][a-z0-9_]*/i],["pun",/^[~!%^&*+=|?:<>/-]/],["lit",/^\b0x[0-9a-f]+/i],["lit",/^\b\d+(?:\.\d*)?(?:e[+-]?\d+)?/i],["lit",
/^\b\.\d+(?:e[+-]?\d+)?/i],["pun",/^[(){}\[\],.;]/]]),["dart"]);
