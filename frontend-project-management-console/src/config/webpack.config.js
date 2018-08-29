const { dev, prod} = require('@ionic/app-scripts/config/webpack.config');
const webpack = require('webpack');
const RetrieveSSMValues = require('./scripts/read-ssm');
const RetrieveLocalConfigValues = require('./scripts/read-local');
const configVarArray = ['auth0-clientId', 'auth0-domain', 'cinco-api-url', 'cla-api-url', 'analytics-api-url'];
const stageEnv = process.env.STAGE_ENV;

module.exports = async env => {
  // Here we hard code stage name, it's not perfect since if a new stage created/modified, we also need to change it.
  const shouldReadFromSSM = (stageEnv !== undefined && stageEnv === 'staging') || (stageEnv !== undefined && stageEnv === 'prod');
  let configMap = {};

  console.log(shouldReadFromSSM);
  
  if (shouldReadFromSSM){
    const profile = 'lf-cla';
    const region = 'us-east-1';
    configMap = await RetrieveSSMValues(configVarArray, stageEnv, region, profile);
  } else {
    configMap = await RetrieveLocalConfigValues(configVarArray);
  }

  const claConfigPlugin = new webpack.DefinePlugin({
    webpackGlobalVars: {
      CLA_API_URL: JSON.stringify(configMap['cla-api-url']),
      CINCO_API_URL: JSON.stringify(configMap['cinco-api-url']),
      ANALYTICS_API_URL: JSON.stringify(configMap['analytics-api-url']),
      AUTH0_DOMAIN: JSON.stringify(configMap['auth0-domain']),
      AUTH0_CLIENT_ID: JSON.stringify(configMap['auth0-clientId'])
    }
  });

  dev.plugins.push(claConfigPlugin);
  prod.plugins.push(claConfigPlugin);

  return {
    dev: dev,
    prod: prod
  };
};
