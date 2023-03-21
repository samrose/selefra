#!/bin/bash
#######################################################################################################################
#                                                                                                                     #
#            This is Selefra's one-click installation script that supports both MacOS and Linux                       #
#                                                                                                                     #
#                                                                                                                     #
#                                                                                                   Version: 0.0.1    #
#                                                                                                                     #
#######################################################################################################################

# ---------------------------------------------------------- init ------------------------------------------------------

RESET="\\033[0m"
RED="\\033[31;1m"
GREEN="\\033[32;1m"
YELLOW="\\033[33;1m"
BLUE="\\033[34;1m"
WHITE="\\033[37;1m"

say_green()
{
    [ -z "${SILENT}" ] && printf "%b%s%b\\n" "${GREEN}" "$1" "${RESET}"
    return 0
}

say_red()
{
    printf "%b%s%b\\n" "${RED}" "$1" "${RESET}"
}

say_yellow()
{
    [ -z "${SILENT}" ] && printf "%b%s%b\\n" "${YELLOW}" "$1" "${RESET}"
    return 0
}

say_blue()
{
    [ -z "${SILENT}" ] && printf "%b%s%b\\n" "${BLUE}" "$1" "${RESET}"
    return 0
}

say_white()
{
    [ -z "${SILENT}" ] && printf "%b%s%b\\n" "${WHITE}" "$1" "${RESET}"
    return 0
}

at_exit()
{
    if [ "$?" -ne 0 ]; then
        >&2 say_red
        >&2 say_red "We're sorry, but it looks like something might have gone wrong during installation."
        >&2 say_red "If you need help, please join us on https://www.selefra.io/community/join"
    fi
}

# TODO Add boot path for source compilation
print_unsupported_platform() {
    >&2 say_red "error: We're sorry, but it looks like selefra is not supported on your platform"
    >&2 say_red "       We support 64-bit versions of Linux and macOS and are interested in supporting"
    >&2 say_red "       more platforms.  Please open an issue at https://github.com/selefra/selefra/issues"
    >&2 say_red "       and let us know what platform you're using!"
}

# get os
OS=""
case $(uname) in
    "Linux") OS="linux";;
    "Darwin") OS="darwin";;
    *)
        print_unsupported_platform
        exit 1
        ;;
esac
say_blue "OS: ${OS}"

# get arch
ARCH=""
case $(uname -m) in
    "x86_64") ARCH="amd64";;
    "arm64") ARCH="arm64";;
    "aarch64") ARCH="arm64";;
    *)
        print_unsupported_platform
        exit 1
        ;;
esac
say_blue "ARCH: ${ARCH}"

# ---------------------------------------------------------- check env -------------------------------------------------

# check wget
which wget
if [ $? -ne 0 ]; then
	say_red "Sorry, you must have wget installed to use this script"
	say_red "wget GNU repo: https://ftp.gnu.org/gnu/wget/"
	exit
fi

# check unzip tools by os
unzip_command=""
file_suffix=""
case $OS in
    "linux")
	# check tar
	which tar
	if [ $? -ne 0 ]; then
	        say_red "Sorry, you must have tar installed to use this script"
	       	say_red "tar homepage: https://www.gnu.org/software/tar/"
	        exit
	fi
	unzip_command="tar zxvf"
	file_suffix=".tar.gz"
	;;
    "darwin")
	# check tar
	which unzip
	if [ $? -ne 0 ]; then
		# TODO Added boot link to install unzip
		say_red "Sorry, you must have unzip installed to use this script"
		exit
	fi
	unzip_command="unzip"
	file_suffix=".zip"
	;;
esac

# ---------------------------------------------------------- download --------------------------------------------------

trap at_exit EXIT

say_blue "begin download selefra..."
# download
download_url="https://github.com/selefra/selefra/releases/latest/download/selefra_${OS}_${ARCH}${file_suffix}"
say_blue "download selefra installation file from $download_url ..."
download_save_path="./selefra_${OS}_${ARCH}${file_suffix}"
wget -t 30 -T 60 $download_url -O $download_save_path
if [ $? -ne 0 ]; then
  say_red "selefra installation file download failed, please check your network and try again!"
  exit
fi
say_green "selefra installation file download success!"

# ---------------------------------------------------------- install ---------------------------------------------------

# unzip
say_blue "begin unzip selefra installation file..."
$unzip_command $download_save_path
say_green "unzip selefra installation file success!"

# Consider it fixed for now, update the policy here if it changes in the future
selefra_executable_file_path="./selefra"

# add $PATH , need to add more path judgment? I don't know...
if [[ $PATH =~ "/usr/local/bin" ]]; then

	copy_to_path="/usr/local/bin/selefra"

	# If it already exists, try to remove it
	if  [ -f "${copy_to_path}" ]; then

		case $OS in
		"linux")
		sudo rm -rf $copy_to_path
		;;
	    	"darwin")
		rm -rf $copy_to_path
		;;
		esac

		if [ $? -ne 0 ]; then
      say_red "The ${copy_to_path} file already exists and cannot be deleted. please manually update the ${copy_to_path}"
		  exit
		fi
	fi

	# Then make a copy of selefra to the target path

	case $OS in
	"linux")
	sudo cp ./selefra /usr/local/bin/selefra
	;;
    	"darwin")
	cp ./selefra /usr/local/bin/selefra
	;;
	esac

	if [ $? -ne 0 ]; then
        	say_red 'copy selefra to $PATH failed. please add it manually'
		exit
	fi

	# Then delete all the files generated during your installation
  say_green "Delete temporary files during installation..."
	rm $download_save_path
  rm $selefra_executable_file_path

fi

# TODO Adds a boot link to the Quick Start document
say_green "selefra download and install success!"


# ----------------------------------------------------------------------------------------------------------------------
