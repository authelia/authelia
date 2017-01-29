/*

 Copyright (C) 2013 Nikhil Dabas

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
PR.registerLangHandler(PR.createSimpleLexer([["pln",/^[\t\n\r \xA0]+/,null,"\t\n\r \u00a0"],["str",/^!?\"(?:[^\"\\]|\\[\s\S])*(?:\"|$)/,null,'"'],["com",/^;[^\r\n]*/,null,";"]],[["pln",/^[%@!](?:[-a-zA-Z$._][-a-zA-Z$._0-9]*|\d+)/],["kwd",/^[A-Za-z_][0-9A-Za-z_]*/,null],["lit",/^\d+\.\d+/],["lit",/^(?:\d+|0[xX][a-fA-F0-9]+)/],["pun",/^[()\[\]{},=*<>:]|\.\.\.$/]]),["llvm","ll"]);
