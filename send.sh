#!/bin/bash

show_menu() {
  title="Pick the lab number [1-6], type 7 to quit:"
  options=("1" "2" "3" "4" "5" "6")

  echo "$title"
  select opt in "${options[@]}" "Quit"; do

      case "$REPLY" in

      [1-6] )
        echo "You picked $opt"
        LAB_NUM=$opt
        break
        ;;

      $(( ${#options[@]}+1 )) ) echo "Goodbye!"; break;;
      *) echo "Invalid option. Try another one.";continue;;

      esac
  done
}

show_menu
lab_path="$HOME/lab"$LAB_NUM
ls $lab_path > /dev/null
if [ $? -ne 0 ] ; then
  echo "there is no $lab_path directory (in $HOME)"
  exit 1
fi

echo "Enter your name in one word [a-zA-Z0-9] : "
read name
echo "Enter group number (m3204/m3205): "
read gn
if [[ $gn =~ ^("m3204"|"m3205")$ ]] ; then
  tar -cf /tmp/my.tar $lab_path
  curl -X POST -F 'data=@/tmp/my.tar' -H "student_name: $name" -H "type: .tar" -H "Content-Type:multipart/form-data" http://35.228.116.28:8080/upload/$gn/lab$LAB_NUM
  rm /tmp/my.tar
  exit 0
fi

echo "wrong group number"