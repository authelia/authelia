/*

 Copyright (C) 2015 Chris Morgan

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
PR.registerLangHandler(PR.createSimpleLexer([],[["pln",/^[\t\n\r \xA0]+/],["com",/^\/\/.*/],["com",/^\/\*[\s\S]*?(?:\*\/|$)/],["str",/^b"(?:[^\\]|\\(?:.|x[\da-fA-F]{2}))*?"/],["str",/^"(?:[^\\]|\\(?:.|x[\da-fA-F]{2}|u\{\[\da-fA-F]{1,6}\}))*?"/],["str",/^b?r(#*)\"[\s\S]*?\"\1/],["str",/^b'([^\\]|\\(.|x[\da-fA-F]{2}))'/],["str",/^'([^\\]|\\(.|x[\da-fA-F]{2}|u\{[\da-fA-F]{1,6}\}))'/],["tag",/^'\w+?\b/],["kwd",/^(?:match|if|else|as|break|box|continue|extern|fn|for|in|if|impl|let|loop|pub|return|super|unsafe|where|while|use|mod|trait|struct|enum|type|move|mut|ref|static|const|crate)\b/],
["kwd",/^(?:alignof|become|do|offsetof|priv|pure|sizeof|typeof|unsized|yield|abstract|virtual|final|override|macro)\b/],["typ",/^(?:[iu](8|16|32|64|size)|char|bool|f32|f64|str|Self)\b/],["typ",/^(?:Copy|Send|Sized|Sync|Drop|Fn|FnMut|FnOnce|Box|ToOwned|Clone|PartialEq|PartialOrd|Eq|Ord|AsRef|AsMut|Into|From|Default|Iterator|Extend|IntoIterator|DoubleEndedIterator|ExactSizeIterator|Option|Some|None|Result|Ok|Err|SliceConcatExt|String|ToString|Vec)\b/],["lit",/^(self|true|false|null)\b/],
["lit",/^\d[0-9_]*(?:[iu](?:size|8|16|32|64))?/],["lit",/^0x[a-fA-F0-9_]+(?:[iu](?:size|8|16|32|64))?/],["lit",/^0o[0-7_]+(?:[iu](?:size|8|16|32|64))?/],["lit",/^0b[01_]+(?:[iu](?:size|8|16|32|64))?/],["lit",/^\d[0-9_]*\.(?![^\s\d.])/],["lit",/^\d[0-9_]*(?:\.\d[0-9_]*)(?:[eE][+-]?[0-9_]+)?(?:f32|f64)?/],["lit",/^\d[0-9_]*(?:\.\d[0-9_]*)?(?:[eE][+-]?[0-9_]+)(?:f32|f64)?/],["lit",/^\d[0-9_]*(?:\.\d[0-9_]*)?(?:[eE][+-]?[0-9_]+)?(?:f32|f64)/],
["atn",/^[a-z_]\w*!/i],["pln",/^[a-z_]\w*/i],["atv",/^#!?\[[\s\S]*?\]/],["pun",/^[+\-/*=^&|!<>%[\](){}?:.,;]/],["pln",/./]]),["rust"]);
