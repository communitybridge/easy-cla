if(process.argv[2] != 'dev') require('newrelic');
var express = require('express');
var passport = require('passport');
var request = require('request');
var multer  = require('multer');
var async = require('async');

var cinco_api = require("../lib/api");

var router = express.Router();

var hostURL = process.env['CINCO_SERVER_URL'];
if(!hostURL.startsWith('http')) hostURL = 'http://' + hostURL;

var cinco = cinco_api(hostURL);

router.get('/mailing/:projectId', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){
    var projectId = req.params.projectId;
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    projManagerClient.getProject(projectId, function (err, project) {
      projManagerClient.getMailingListsAndParticipants(projectId, function (err, mailingLists) {
        res.render('mailing', { mailingLists: mailingLists, project:project });
      });
    });
  }
});

router.post('/mailing/:projectId', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    var projectId = req.params.projectId;

    var mailingName = req.body.mailing_name;
    var mailingEmailAdmin = req.body.mailing_email_admin;
    var mailingPassword = req.body.mailing_password;
    var mailingSubscribePolicy = req.body.subscribe_policy_radio;
    var mailingArchivePolicy = req.body.archive_policy_radio;

    var newMailingList = {
      "name": mailingName,
      "admin": mailingEmailAdmin,
      "password": mailingPassword,
      "subscribePolicy": mailingSubscribePolicy,
      "archivePolicy": mailingArchivePolicy
    };

    projManagerClient.createMailingList(projectId, newMailingList, function (err, created, mailingListId) {
      console.log("mailing list created: " + created);
      console.log("mailingListId: " + mailingListId);
      return res.redirect('/mailing/' + projectId);
    });
  }
});

router.post('/addParticipantToMailingList/:projectId/', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    var projectId = req.params.projectId;
    var mailinglistId = req.body.mailing_list_id;
    var participant_email = req.body.participant_email;
    var newParticipant = {
      "address": participant_email
    };
    projManagerClient.addParticipantToMailingList(projectId, mailinglistId, newParticipant, function (err, created, response) {
      return res.redirect('/mailing/' + projectId);
    });
  }
});

module.exports = router;
