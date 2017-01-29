/*

 Copyright (C) 2013 Andrew Allen

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
PR.registerLangHandler(PR.createSimpleLexer([["pln",/^[\t\n\x0B\x0C\r ]+/,null,"\t\n\x0B\f\r "],["str",/^\"(?:[^\"\\\n\x0C\r]|\\[\s\S])*(?:\"|$)/,null,'"'],["lit",/^[a-z][a-zA-Z0-9_]*/],["lit",/^\'(?:[^\'\\\n\x0C\r]|\\[^&])+\'?/,null,"'"],["lit",/^\?[^ \t\n({]+/,null,"?"],["lit",/^(?:0o[0-7]+|0x[\da-f]+|\d+(?:\.\d+)?(?:e[+\-]?\d+)?)/i,null,"0123456789"]],[["com",/^%[^\n]*/],["kwd",/^(?:module|attributes|do|let|in|letrec|apply|call|primop|case|of|end|when|fun|try|catch|receive|after|char|integer|float,atom,string,var)\b/],
["kwd",/^-[a-z_]+/],["typ",/^[A-Z_][a-zA-Z0-9_]*/],["pun",/^[.,;]/]]),["erlang","erl"]);
