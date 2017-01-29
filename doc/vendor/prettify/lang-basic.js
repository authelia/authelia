/*

 Copyright (C) 2013 Peter Kofler

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
PR.registerLangHandler(PR.createSimpleLexer([["str",/^(?:"(?:[^\\"\r\n]|\\.)*(?:"|$))/,null,'"'],["pln",/^\s+/,null," \r\n\t\u00a0"]],[["com",/^REM[^\r\n]*/,null],["kwd",/^\b(?:AND|CLOSE|CLR|CMD|CONT|DATA|DEF ?FN|DIM|END|FOR|GET|GOSUB|GOTO|IF|INPUT|LET|LIST|LOAD|NEW|NEXT|NOT|ON|OPEN|OR|POKE|PRINT|READ|RESTORE|RETURN|RUN|SAVE|STEP|STOP|SYS|THEN|TO|VERIFY|WAIT)\b/,null],["pln",/^[A-Z][A-Z0-9]?(?:\$|%)?/i,null],["lit",/^(?:\d+(?:\.\d*)?|\.\d+)(?:e[+\-]?\d+)?/i,
null,"0123456789"],["pun",/^.[^\s\w\.$%"]*/,null]]),["basic","cbm"]);
