#apt install whiptail -y

whiptail \
    --title "Authelia Theme Installer" \
    --msgbox "This will install a theme to authelia." 10 60

MODE=$(whiptail --title "Which mode?" --radiolist \
"Which mode?" 10 60 2 \
"local" "Local install" ON \
"global" "Global install" OFF  3>&1 1>&2 2>&3)

exitstatus=$?
if [ $exitstatus = 1 ];
then
	exit 1
fi

THEME=$(whiptail --title "Which theme?" --radiolist \
"Which theme?" 10 60 5 \
"default" "Default theme" ON \
"matrix" "Matrix style theme" OFF \
"black" "Black theme" OFF \
"triangles" "Theme with triangles" OFF \
"squares" "Theme with squares" OFF 3>&1 1>&2 2>&3)

exitstatus=$?
if [ $exitstatus = 1 ];
then
	exit 1
fi

if (whiptail --title "Verbose?" --yesno "Verbose output mode?" 10 60 3>&1 1>&2 2>&3) then
    VERBOSE="yes"
else
    VERBOSE="no"
fi

if (whiptail --title "Build theme?" --yes-button "Build" --no-button "Copy" --yesno "Build theme or copy prebuilt?" 10 60 3>&1 1>&2 2>&3) then
    BUILD="build"
else
    BUILD="copy"
fi

until [[ $PORT =~ ^(0|[1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$ ]]; do
        PORT=$(whiptail --inputbox "Enter the listening port" 10 60 8080 --title "Which port should be exposed?" 10 60 3>&1 1>&2 2>&3)
	exitstatus=$?
	if [ $exitstatus = 1 ];
	then
		exit 1
	fi
done

if (whiptail --title "Authelia Theme Installer" --yes-button "Ok" --no-button "Cancel" --yesno "Summary: \n Mode: $MODE \n Theme: $THEME \n Verbose: $VERBOSE \n Port: $PORT \n Build: $BUILD" 13 70 3>&1 1>&2 2>&3) then
	GO="yes"
else
	exit 1
fi

if test "$VERBOSE" = 'yes'
then
        if test "$BUILD" = 'build'
        then
                #./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT -v -b
                COMMAND="./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT -v -b"
        else
                COMMAND="./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT -v"
        fi
else
        if test "$BUILD" = 'copy'
        then
                COMMAND="./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT"
        else
                COMMAND="./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT -b"
        fi
fi

echo $COMMAND
eval $COMMAND

#whiptail --title "Example" --gauge "Just another example" 6 50 0

#if test "$VERBOSE" = 'yes'
#then
#	if test "$BUILD" = 'build'
#	then
#		#./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT -v -b
#		COMMAND="./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT -v -b"
#	else
#		COMMAND="./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT -v"
#	fi
#else
#	if test "$BUILD" = 'copy'
#        then
#                COMMAND="./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT"
#        else
#                COMMAND="./authelia-theme-install_verbose.sh -t $THEME -m $MODE -p $PORT -b"
#        fi
#fi
