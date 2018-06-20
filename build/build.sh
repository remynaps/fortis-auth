
if [ "$1" == "-linux" ]; then
  export GOOS=linux
  echo "building for Linux.."
else
  echo "building for OSX.."
fi

# small script that interated over the cmd folder
cd $GOPATH/src/github.com/remynaps/fortis/models
echo "building models package"
go build

cd $GOPATH/src/github.com/remynaps/fortis/authorization
echo "building authorization package"
go build

cd $GOPATH/src/github.com/remynaps/fortis/

# find the binaries in the cmd folder
for CMD in `ls cmd`; do
  echo "found executable: " $CMD
  # cd into the folder and run go build
  cd $GOPATH/src/github.com/remynaps/fortis/cmd/$CMD/
  go build -o $GOPATH/src/github.com/remynaps/fortis/bin/$CMD
done

export GOOS=darwin

echo "Done!"