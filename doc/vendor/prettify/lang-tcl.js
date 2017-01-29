/*

 Copyright (C) 2012 Pyrios

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
PR.registerLangHandler(PR.createSimpleLexer([["opn",/^\{+/,null,"{"],["clo",/^\}+/,null,"}"],["com",/^#[^\r\n]*/,null,"#"],["pln",/^[\t\n\r \xA0]+/,null,"\t\n\r \u00a0"],["str",/^\"(?:[^\"\\]|\\[\s\S])*(?:\"|$)/,null,'"']],[["kwd",/^(?:after|append|apply|array|break|case|catch|continue|error|eval|exec|exit|expr|for|foreach|if|incr|info|proc|return|set|switch|trace|uplevel|upvar|while)\b/,null],["lit",/^[+\-]?(?:[0#]x[0-9a-f]+|\d+\/\d+|(?:\.\d+|\d+(?:\.\d*)?)(?:[ed][+\-]?\d+)?)/i],
["lit",/^\'(?:-*(?:\w|\\[\x21-\x7e])(?:[\w-]*|\\[\x21-\x7e])[=!?]?)?/],["pln",/^-*(?:[a-z_]|\\[\x21-\x7e])(?:[\w-]*|\\[\x21-\x7e])[=!?]?/i],["pun",/^[^\w\t\n\r \xA0()\"\\\';]+/]]),["tcl"]);
