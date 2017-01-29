/*

 Copyright (C) 2013 Eric Knibbe

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
PR.registerLangHandler(PR.createSimpleLexer([["pln",/^[\t\n\r \xA0]+/,null,"\t\n\r \u00a0"],["str",/^\'(?:[^\'\\]|\\[\s\S])*(?:\'|$)/,null,"'"],["str",/^\"(?:[^\"\\]|\\[\s\S])*(?:\"|$)/,null,'"'],["str",/^\`[^\`]*(?:\`|$)/,null,"`"],["lit",/^0x[\da-f]+|\d+/i,null,"0123456789"],["atn",/^#\d+|[#$][a-z_][\w.]*|#![ \S]+lasso9\b/i,null,"#$"]],[["tag",/^[[\]]|<\?(?:lasso(?:script)?|=)|\?>|noprocess\b|no_square_brackets\b/i],["com",/^\/\/[^\r\n]*|\/\*[\s\S]*?\*\//],
["atn",/^-(?!infinity)[a-z_][\w.]*|\.\s*'[a-z_][\w.]*'/i],["lit",/^\d*\.\d+(?:e[-+]?\d+)?|infinity\b|NaN\b/i],["atv",/^::\s*[a-z_][\w.]*/i],["lit",/^(?:true|false|none|minimal|full|all|void|and|or|not|bw|nbw|ew|new|cn|ncn|lt|lte|gt|gte|eq|neq|rx|nrx|ft)\b/i],["kwd",/^(?:error_code|error_msg|error_pop|error_push|error_reset|cache|database_names|database_schemanames|database_tablenames|define_tag|define_type|email_batch|encode_set|html_comment|handle|handle_error|header|if|inline|iterate|ljax_target|link|link_currentaction|link_currentgroup|link_currentrecord|link_detail|link_firstgroup|link_firstrecord|link_lastgroup|link_lastrecord|link_nextgroup|link_nextrecord|link_prevgroup|link_prevrecord|log|loop|namespace_using|output_none|portal|private|protect|records|referer|referrer|repeating|resultset|rows|search_args|search_arguments|select|sort_args|sort_arguments|thread_atomic|value_list|while|abort|case|else|if_empty|if_false|if_null|if_true|loop_abort|loop_continue|loop_count|params|params_up|return|return_value|run_children|soap_definetag|soap_lastrequest|soap_lastresponse|tag_name|ascending|average|by|define|descending|do|equals|frozen|group|handle_failure|import|in|into|join|let|match|max|min|on|order|parent|protected|provide|public|require|returnhome|skip|split_thread|sum|take|thread|to|trait|type|where|with|yield|yieldhome)\b/i],
["typ",/^(?:array|date|decimal|duration|integer|map|pair|string|tag|xml|null|boolean|bytes|keyword|list|locale|queue|set|stack|staticarray|local|var|variable|global|data|self|inherited|currentcapture|givenblock)\b|^\.\.?/i],["pln",/^[a-z_][\w.]*(?:=\s*(?=\())?/i],["pun",/^:=|[-+*\/%=<>&|!?\\]/]]),["lasso","ls","lassoscript"]);
