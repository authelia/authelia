In order to build a specific Theme you need to run:
<br>`grunt --theme=<theme_name>`

Available themes are: default, black, matrix, squares, triangles:
<br>`grunt --theme=black`

By default the original theme will be built.

If you want to create a new theme:
- Use the `themes/src/default` or `themes/full/default` as source material
- Make a copy in themes folder with a new name
- Add your theme folder name on line 237,239 and 242
- And then build as above, with your theme folder/name.

`<src>/<theme_name>` contains the source files, before build.  
`<full>/<theme_name>` contains the pre-built files.

authelia-theme-install.sh is meant for npm install, either locally (/tmp) or globally.
                                                                       
Default usage:                                                         
`authelia-theme-install.sh -i | --interactive`
                                                                       
or adding parameters to default usage:                                 
   -t or --theme <default|black|matrix|squares|triangles>              
   -m or --mode <local|global>                                         
   -p or --port <port number>  
   -v or --verbose

Example:<br>`chmod +x authelia-theme-install.sh && ./authelia-theme-install.sh -t black -m global -p 88 -v`
