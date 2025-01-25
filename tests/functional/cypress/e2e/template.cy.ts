import {validateApiResponse,validate_200_Status,getTokenKey} from '../support/commands'

describe("To Validate & get list of template via API call", function () {
  
 // Define a variable for the environment
 const environment = Cypress.env("CYPRESS_ENV");

 // Import the appropriate configuration based on the environment
 let appConfig;
 if (environment === 'dev') {
   appConfig = require('../appConfig/config.dev.ts').appConfig;
 } else if (environment === 'production') {
   appConfig = require('../appConfig/config.production.ts').appConfig;
 }
  
  //Reference api doc: https://api-gw.dev.platform.linuxfoundation.org/cla-service/v4/api-docs#tag/template
   const claEndpoint = `${Cypress.env("APP_URL")}cla-service/v4`;
    const claGroupId=appConfig.claGroupId; //project name: easyAutom foundation
    let templateId="";

    let bearerToken: string = null;
    before(() => { 
        if(bearerToken==null){
        getTokenKey(bearerToken);
        cy.window().then((win) => {
            bearerToken = win.localStorage.getItem('bearerToken');
          });
        }
    });

it("Endpoint to return the list of available templates", function () {
    cy.request({
      method: 'GET',
      url: `${claEndpoint}/template`,
      auth: {
        'bearer': bearerToken,
      }
    }).then((response) => {
      validate_200_Status(response);
      let list=response.body;
      templateId=list[0].ID;
      expect(list[0].Name).to.eql('Apache Style');
      expect(list[1].Name).to.eql('ASWF 2020 v2.1');
          //To validate schema of response
             validateApiResponse("templates/getTemplates.json",response)
        });
  });

it("Create new templates for a CLA Group", function () {
    cy.request({
      method: 'POST',
      url: `${claEndpoint}/clagroup/${claGroupId}/template`,
      auth: {
        'bearer': bearerToken,
      },body:{
        "MetaFields": [
          {
          "description": "Project's Full Name.",
          "name": "Project Name",
          "templateVariable": "PROJECT_NAME",
          "value" : "Test"
          },
          {
              "description": "The Full Entity Name of the Project.",
              "name": "Project Entity Name",
              "templateVariable": "PROJECT_ENTITY_NAME",
              "value" : "Test"
          },
          {
              "description": "The E-Mail Address of the Person managing the CLA. ",
              "name": "Contact Email Address",
              "templateVariable": "CONTACT_EMAIL",
              "value" : "veerendrat@proximabiz.com"
          }                       
       ],
        "TemplateID": templateId
      },
      failOnStatusCode: false
    }).then((response) => {
      validate_200_Status(response);  
      expect(response.body.corporatePDFURL);
      expect(response.body.individualPDFURL);
        });
  });

it("Preview ICLA templates for CLA Group", function () {
    cy.request({
      method: 'POST',
      url: `${claEndpoint}/template/preview?template_for=icla`,
      auth: {
        'bearer': bearerToken,
      },body:{
        "MetaFields": [
          {
          "description": "Project's Full Name.",
          "name": "Project Name",
          "templateVariable": "PROJECT_NAME",
          "value" : "Test"
          },
          {
              "description": "The Full Entity Name of the Project.",
              "name": "Project Entity Name",
              "templateVariable": "PROJECT_ENTITY_NAME",
              "value" : "Test"
          },
          {
              "description": "The E-Mail Address of the Person managing the CLA. ",
              "name": "Contact Email Address",
              "templateVariable": "CONTACT_EMAIL",
              "value" : "veerendrat@proximabiz.com"
          }                       
       ],
        "TemplateID": templateId
      },
      failOnStatusCode: false
    }).then((response) => {
      validate_200_Status(response);  
      expect(response.body.corporatePDFURL);
      expect(response.body.individualPDFURL);  
        });
  });

it("Preview CCLA templates for CLA Group", function () {
    cy.request({
      method: 'POST',
      url: `${claEndpoint}/template/preview?template_for=ccla`,
      auth: {
        'bearer': bearerToken,
      },body:{
        "MetaFields": [
          {
          "description": "Project's Full Name.",
          "name": "Project Name",
          "templateVariable": "PROJECT_NAME",
          "value" : "Test"
          },
          {
              "description": "The Full Entity Name of the Project.",
              "name": "Project Entity Name",
              "templateVariable": "PROJECT_ENTITY_NAME",
              "value" : "Test"
          },
          {
              "description": "The E-Mail Address of the Person managing the CLA. ",
              "name": "Contact Email Address",
              "templateVariable": "CONTACT_EMAIL",
              "value" : "veerendrat@proximabiz.com"
          }                       
       ],
        "TemplateID": templateId
      },
      failOnStatusCode: false
    }).then((response) => {
      validate_200_Status(response);  
      expect(response.body.corporatePDFURL);
      expect(response.body.individualPDFURL);  
        });
  });

it("Preview CLA Group Template PDF", function () {
    cy.request({
      method: 'GET',
      url: `${claEndpoint}/template/${claGroupId}/preview?claType=ccla&watermark=false`,
      auth: {
        'bearer': bearerToken,
      }
    }).then((response) => {
      validate_200_Status(response);
        });
  });

})