
#!/usr/bin/env bash
set -e

usage () {
  echo "Usage : $0 -s <stage> -r <region of api> [-c](enable cloudfront)"
}

# Get STAGE and CLOUDFRONT configuration from command line.
CLOUDFRONT=false
while getopts ":s:r:c" opts; do
  case ${opts} in
    s) STAGE=${OPTARG} ;;
    r) REGION=${OPTARG} ;;
    c) CLOUDFRONT=true ;;
    *) break ;;
  esac
done
# Removes the parsed command line opts
shift $((OPTIND-1))

if [ -z "${STAGE}" ]; then
  usage
  exit 1
fi

if [ -z "${REGION}" ]; then
  usage
  exit 1
fi

echo 'Building Distribution'
cd src
../node_modules/.bin/ionic build
cd ../

echo 'Building Edge Function'
cd edge
yarn build
cd ../

echo 'Deploying Cloudfront and lambda@edge'
sls deploy --stage="${STAGE}" --cloudfront="${CLOUDFRONT}"

echo 'Deploying Frontend Bucket'
sls client deploy --stage="${STAGE}" --cloudfront="${CLOUDFRONT}" --no-confirm --no-policy-change --no-config-change

if [ ${CLOUDFRONT} = true ]; then
  echo 'Invalidating Cloudfront'
  sls invalidate --stage="${STAGE}" --cloudfront="${CLOUDFRONT}"
fi

exit 0