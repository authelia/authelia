In order to build a specific Theme you need to run:

grunt --theme=<theme_name>

Available themes are: default, black, matrix, squares, triangles

Ex. grunt --theme=black

By default the original theme will be built.

If you want to create a new theme:
- Use the themes/default as source material
- Make a copy in themes folder with a new name
- Add your theme folder name on line 237,239 and 242
- And then build as above, with your theme folder/name.

<src>/<theme_name> contains the source files, before build.
<full>/<theme_name> contains the built files.

authelia-theme-install.sh is meant for npm install, either locally (temp) or globally.
usage is just running the script or giving parameters:
-t or --theme <theme_name>
-m or --mode <local|global>

Ex. chmod +x authelia-theme-install.sh && ./authelia-theme-install.sh -t black -m global

That's it!
