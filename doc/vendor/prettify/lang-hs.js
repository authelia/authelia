/*

 Copyright (C) 2009 Google Inc.

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
PR.registerLangHandler(PR.createSimpleLexer([["pln",/^[\t\n\x0B\x0C\r ]+/,null,"\t\n\x0B\f\r "],["str",/^\"(?:[^\"\\\n\x0C\r]|\\[\s\S])*(?:\"|$)/,null,'"'],["str",/^\'(?:[^\'\\\n\x0C\r]|\\[^&])\'?/,null,"'"],["lit",/^(?:0o[0-7]+|0x[\da-f]+|\d+(?:\.\d+)?(?:e[+\-]?\d+)?)/i,null,"0123456789"]],[["com",/^(?:(?:--+(?:[^\r\n\x0C]*)?)|(?:\{-(?:[^-]|-+[^-\}])*-\}))/],["kwd",/^(?:case|class|data|default|deriving|do|else|if|import|in|infix|infixl|infixr|instance|let|module|newtype|of|then|type|where|_)(?=[^a-zA-Z0-9\']|$)/,
null],["pln",/^(?:[A-Z][\w\']*\.)*[a-zA-Z][\w\']*/],["pun",/^[^\t\n\x0B\x0C\r a-zA-Z0-9\'\"]+/]]),["hs"]);
