#!/bin/bash

# Check if user is root or sudo
if ! [ $(id -u) = 0 ]; then echo "Please run this script as sudo or root"; exit 1 ; fi

#Authelia Requirements
authelia_req=('wget' 'unzip' 'nginx' 'nodejs' 'curl')
authelia_reqname=('WGET' 'UNZIP' 'NGINX' 'NODEJS' 'CURL')
node_debian=('curl -sL https://deb.nodesource.com/setup_8.x | bash - && apt-get install -y nodejs')
interactive=""

# Colors to use for output
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
GREEN='\033[0;32m'
LIGHTBLUE='\033[1;36m'
NC='\033[0m' # No Color

# Get script arguments for non-interactive mode

while [ "$1" != "" ] || [ "$2" != "" ] || [ "$3" != "" ] || [ "$4" != "" ]; do
    case $1 in
        -t | --theme )
            shift
            theme="$1"
            interactive="yes"
            ;;
    esac
    case $1 in
        -m | --mode )
            shift
            mode="$1"
            interactive="yes"
            ;;
    esac
    case $1 in
        -p | --port )
            shift
            port="$1"
            interactive="yes"
            ;;
    esac
    case $1 in
        -h | --help )
            shift
            echo -e "${LIGHTBLUE}--------------------------------------------------------------------------------------------"
            echo -e "${LIGHTBLUE}| authelia-theme-install.sh is meant for npm install, either locally (/tmp) or globally.   |"
            echo -e "${LIGHTBLUE}|                                                                                          |"
            echo -e "${LIGHTBLUE}| Default usage:                                                                           |"
            echo -e "${LIGHTBLUE}| authelia-theme-install.sh -i | --interactive                                             |"
            echo -e "${LIGHTBLUE}|                                                                                          |"
            echo -e "${LIGHTBLUE}| or adding parameters to default usage:                                                   |"
            echo -e "${LIGHTBLUE}|  -t or --theme <default|black|matrix|squares|triangles>                                  |"
            echo -e "${LIGHTBLUE}|  -m or --mode <local|global>                                                             |"
            echo -e "${LIGHTBLUE}|  -p or --port <port number>                                                              |"
            echo -e "${LIGHTBLUE}--------------------------------------------------------------------------------------------${NC}"
            exit 0
            ;;
    esac
    case $1 in
        -i | --interactive )
            shift
            interactive="yes"
            ;;
    esac
    case $2 in
        -t | --theme )
            shift
            theme="$2"
            interactive="yes"
            ;;
    esac
    case $2 in
        -m | --mode )
            shift
            mode="$2"
            interactive="yes"
            ;;
    esac
    case $2 in
        -p | --port )
            shift
            port="$2"
            interactive="yes"
            ;;
    esac
    case $3 in
        -t | --theme )
            shift
            theme="$3"
            interactive="yes"
            ;;
    esac
    case $3 in
        -m | --mode )
            shift
            mode="$3"
            interactive="yes"
            ;;
    esac
    case $3 in
        -p | --port )
            shift
            port="$3"
            interactive="yes"
            ;;
    esac
    shift
done

if test "$interactive" != 'yes'
then
    echo -e "${RED}For interactive mode please use -i, or use -h | --help${NC}"
    exit 0
fi

#echo "port: "$port
#echo "interactive: "$interactive
#echo "theme: "$theme
#echo "mode: "$mode

authelia_mod()
	{
        echo
        echo -e "${LIGHTBLUE}> Updating apt repositories...\e[0m${NC}"
		apt-get update >/dev/null 2>&1
        echo
		echo -e "${LIGHTBLUE}> Adding nodejs...\e[0m${NC}"
		apt-get -y install ${node_debian} >/dev/null 2>&1
		for ((i=0; i < "${#authelia_reqname[@]}"; i++))
		do
		    echo -e "${LIGHTBLUE}> Installing ${authelia_reqname[$i]}...\e[0m${NC}"
		    apt-get -y install ${authelia_req[$i]} >/dev/null 2>&1
		done
		echo
    }

authelia_mod

dest_global="$(echo | npm -g root)"
dest_local="$(echo | pwd)"

authelia_local_install()
	{
	while [[ "$theme" != 'default' && "$theme" != 'black' && "$theme" != 'matrix' && "$theme" != 'squares' && "$theme" != 'triangles' ]]; do
        echo -e "${YELLOW}> Which theme? ([default],black,matrix,triangles,squares)${NC}"
        read theme
            if test -z "$theme"
            then
            	theme="default"
                echo -e "${LIGHTBLUE}> Input empty, defaulting to:" $theme"...${NC}"

		echo -e "${LIGHTBLUE}> Installing latest Authelia locally...${NC}"

                rm -rf /tmp/authelia
                mkdir /tmp/authelia && cd /tmp/authelia

                #authelia_latest_github="$(curl -s https://github.com/clems4ever/authelia/releases | grep authelia/archive/v* | cut -d '"' -f 4 | head -n 2 | grep tar)"
                #authelia_version_github=${authelia_latest_github##*/}
                #curl -s -OL https://github.com/$authelia_latest_github

                #authelia_latest_tarball="$(npm view authelia | grep tgz | cut -d " " -f 2)"
                #authelia_version_npm=${authelia_latest_tarball##*/}
                #curl -s -OL $authelia_latest_tarball

                echo -e "${LIGHTBLUE}> Cloning git...${NC}"
                git clone --single-branch --branch dev https://github.com/bankainojutsu/authelia.git /tmp/authelia >/dev/null 2>&1 && cd /tmp/authelia

                echo -e "${LIGHTBLUE}> Getting latest tarball...${NC}"
                authelia_latest_npm="$(echo | npm view authelia dist.tarball)"
                authelia_filename_npm=${authelia_latest_npm##*/}
                authelia_filename="$(echo $authelia_filename_npm)"
                curl -s -OL $(npm view authelia dist.tarball)
                #npm pack authelia

                tar -zxf $authelia_filename

                echo -e "${LIGHTBLUE}> Installing...${NC}"
                npm install >/dev/null 2>&1

                echo -e "${LIGHTBLUE}> Copying $theme...${NC}"
                cp -R "./themes/full/$theme/public_html/" "./package/dist/server/src/"
                cp -R "./themes/full/$theme/resources/" "./package/dist/server/src/"
                cp -R "./themes/full/$theme/views/" "./package/dist/server/src/"

                if test -z "$port"
                then
                    echo "Using port: 8080"
                else
                    echo "Using port: "$port
                    sed -i 's6/8080/'$port'/g' ./themes/config.minimal.port.yml
                fi

                node package/dist/server/src/index.js ./themes/config.minimal.port.yml
            else
                echo -e "${LIGHTBLUE}> Using theme:" $theme"...${NC}"

                echo -e "${LIGHTBLUE}> Installing latest Authelia locally...${NC}"

                rm -rf /tmp/authelia
		mkdir /tmp/authelia && cd /tmp/authelia

                #authelia_latest_github="$(curl -s https://github.com/clems4ever/authelia/releases | grep authelia/archive/v* | cut -d '"' -f 4 | head -n 2 | grep tar)"
                #authelia_version_github=${authelia_latest_github##*/}
                #curl -s -OL https://github.com/$authelia_latest_github

                #authelia_latest_tarball="$(npm view authelia | grep tgz | cut -d " " -f 2)"
                #authelia_version_npm=${authelia_latest_tarball##*/}
                #curl -s -OL $authelia_latest_tarball

                echo -e "${LIGHTBLUE}> Cloning git...${NC}"
                git clone --single-branch --branch dev https://github.com/bankainojutsu/authelia.git /tmp/authelia >/dev/null 2>&1 && cd /tmp/authelia

                echo -e "${LIGHTBLUE}> Getting latest tarball...${NC}"
                authelia_latest_npm="$(echo | npm view authelia dist.tarball)"
                authelia_filename_npm=${authelia_latest_npm##*/}
                authelia_filename="$(echo $authelia_filename_npm)"
                curl -s -OL $(npm view authelia dist.tarball)
                #npm pack authelia

                tar -zxf $authelia_filename

                echo -e "${LIGHTBLUE}> Installing...${NC}"
                npm install >/dev/null 2>&1

                echo -e "${LIGHTBLUE}> Copying $theme...${NC}"
                cp -R "./themes/full/$theme/public_html/" "./package/dist/server/src/"
                cp -R "./themes/full/$theme/resources/" "./package/dist/server/src/"
                cp -R "./themes/full/$theme/views/" "./package/dist/server/src/"

                if test -z "$port"
                then
                    echo "Using port: 8080"
                else
                    echo "Using port: "$port
                    sed -i 's6/8080/'$port'/g' ./themes/config.minimal.port.yml
                fi

                node package/dist/server/src/index.js ./themes/config.minimal.port.yml
            fi
    done
    while [[ "$theme" = 'default' || "$theme" = 'black' || "$theme" = 'matrix' || "$theme" = 'squares' || "$theme" = 'triangles' ]]; do
                echo -e "${LIGHTBLUE}> Using theme:" $theme"...${NC}"

                echo -e "${LIGHTBLUE}> Installing latest Authelia locally...${NC}"

                rm -rf /tmp/authelia
                mkdir /tmp/authelia && cd /tmp/authelia

                #authelia_latest_github="$(curl -s https://github.com/clems4ever/authelia/releases | grep authelia/archive/v* | cut -d '"' -f 4 | head -n 2 | grep tar)"
                #authelia_version_github=${authelia_latest_github##*/}
                #curl -s -OL https://github.com/$authelia_latest_github

                #authelia_latest_tarball="$(npm view authelia | grep tgz | cut -d " " -f 2)"
                #authelia_version_npm=${authelia_latest_tarball##*/}
                #curl -s -OL $authelia_latest_tarball

                echo -e "${LIGHTBLUE}> Cloning git...${NC}"
                git clone --single-branch --branch dev https://github.com/bankainojutsu/authelia.git /tmp/authelia >/dev/null 2>&1 && cd /tmp/authelia

                echo -e "${LIGHTBLUE}> Getting latest tarball...${NC}"
                authelia_latest_npm="$(echo | npm view authelia dist.tarball)"
                authelia_filename_npm=${authelia_latest_npm##*/}
                authelia_filename="$(echo $authelia_filename_npm)"
                curl -s -OL $(npm view authelia dist.tarball)
                #npm pack authelia

                tar -zxf $authelia_filename

                echo -e "${LIGHTBLUE}> Installing...${NC}"
                npm install >/dev/null 2>&1

                echo -e "${LIGHTBLUE}> Copying $theme...${NC}"
                cp -R "./themes/full/$theme/public_html/" "./package/dist/server/src/"
                cp -R "./themes/full/$theme/resources/" "./package/dist/server/src/"
                cp -R "./themes/full/$theme/views/" "./package/dist/server/src/"

                if test -z "$port"
                then
                    echo "Using port: 8080"
                else
                    echo "Using port: "$port
                    sed -i 's6/8080/'$port'/g' ./themes/config.minimal.port.yml
                fi

                node package/dist/server/src/index.js ./themes/config.minimal.port.yml
    done
	}

authelia_global_install()
	{
echo -e "${LIGHTBLUE}> Installing Authelia globally...${NC}"
npm install -g authelia >/dev/null 2>&1

echo -e "${LIGHTBLUE}> Installing Grunt-cli globally...${NC}"
npm install -g grunt-cli >/dev/null 2>&1

echo -e "${LIGHTBLUE}> Creating user Authelia...${NC}"
useradd -r -s /bin/false authelia >/dev/null 2>&1

echo -e "${LIGHTBLUE}> Configuring Authelia...${NC}"
mkdir -p /etc/authelia
chown root:authelia /etc/authelia
chmod 2750 /etc/authelia
wget -O /etc/authelia/config.yml 'https://raw.githubusercontent.com/clems4ever/authelia/master/config.template.yml' >/dev/null 2>&1
wget -O /etc/authelia/config.minimal.yml 'https://raw.githubusercontent.com/clems4ever/authelia/master/config.minimal.yml' >/dev/null 2>&1
wget -O /etc/authelia/users_database.yml 'https://raw.githubusercontent.com/clems4ever/authelia/master/users_database.yml' >/dev/null 2>&1

echo -e "${LIGHTBLUE}> Creating Authelia in systemd...${NC}"
cat >/etc/systemd/system/authelia.service <<EOL
[Unit]
Description=2FA Single Sign-On Authentication Server
After=network.target

[Service]
User=authelia
Group=authelia
ExecStart=/usr/bin/authelia /etc/authelia/config.minimal.yml
Restart=always

[Install]
WantedBy=multi-user.target
EOL

echo -e "${LIGHTBLUE}> Reload daemon, enable and start service...${NC}"
systemctl daemon-reload
systemctl start authelia
systemctl status authelia
systemctl enable authelia

echo -e "${LIGHTBLUE}> Deleting /tmp/authelia...${NC}"
rm -rf /tmp/authelia

echo -e "${LIGHTBLUE}> Cloning git...${NC}"
git clone https://github.com/bankainojutsu/authelia.git /tmp/authelia >/dev/null 2>&1 && cd /tmp/authelia

if test -z "$theme"
then
	while [[ "$theme" != 'default' && "$theme" != 'black' && "$theme" != 'matrix' && "$theme" != 'squares' && "$theme" != 'triangles' ]]; do
		echo -e "${YELLOW}> Which theme? ([default],black,matrix,triangles,squares)${NC}"
		read theme
		if test -z "$theme"
		then
            theme="default"
			echo -e "${LIGHTBLUE}> Input empty, defaulting to:" $theme"...${NC}"
		else
			echo
		fi
	done
else
		echo -e "${LIGHTBLUE}> Theme chosen:" $theme"...${NC}"
fi

echo -e "${LIGHTBLUE}> Building theme:" $theme"...${NC}"
npm install >/dev/null 2>&1 && echo -e "${YELLOW}50%...${NC}"

grunt --theme=$theme >/dev/null 2>&1 && echo -e "${GREEN}100%... OK!${NC}"

echo -e "${LIGHTBLUE}> Installing" $theme "theme${NC}"
cp -r dist $dest"/authelia" 

echo -e "${LIGHTBLUE}> Starting server...${NC}"
echo -e "${LIGHTBLUE}> Stop with CTRL-C, run with \"authelia config.file\"${NC}"

sleep 5
systemctl status authelia
	}

authelia_global_or_local_install()
	{
	while [ "$mode" != "local" ] && [ "$mode" != "global" ]; do    
        echo -e "${YELLOW}> global or [local]?${NC}"
        read mode
		if test -z "$mode"
		then
            mode="local"
			echo -e "${LIGHTBLUE}> Input empty, defaulting to:" $mode"...${NC}"
			authelia_local_install
        fi
    done
        if test "$mode" = 'local'
		then
			echo -e "${LIGHTBLUE}> "$mode" install...${NC}"
			authelia_local_install
		elif test "$mode" = 'global'
		then
			echo -e "${LIGHTBLUE}> "$mode" install...${NC}"
			authelia_global_install
		fi
	}

authelia_global_or_local_install

#cp config.minimal.yml /etc/authelia && rm -rf /tmp/authelia && cd /home && authelia /etc/authelia/config.minimal.yml
