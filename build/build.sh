
if [ "$1" == "-linux" ]; then
  export GOOS=linux
  echo "building for Linux.."
else
  echo "building for OSX.."
fi

# small script that interated over the cmd folder

# database package
cd models
go build

cd ../ # root project path

# authorization package
cd authorization
go build

cd ../ # root project path

# find the binaries in the cmd folder
for CMD in `ls cmd`; do
  echo "found executable: " $CMD
  # cd into the folder and run go build
  cd ./cmd/$CMD/
  go build -o ../../bin/$CMD
  cd ../../
done

export GOOS=darwin

echo "Done!"