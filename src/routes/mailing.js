if (process.env['NEWRELIC_LICENSE']) require('newrelic');
var express = require('express');
var router = express.Router();

var cinco = require("../lib/api");

/**
* Mailing Lists
* Resources for working with mailing lists of projects
**/

router.get('/mailing/:projectId', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){
    var projectId = req.params.projectId;
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    projManagerClient.getProject(projectId, function (err, project) {
      projManagerClient.getMailingListsAndParticipants(projectId, function (err, mailingLists) {
        res.send(mailingLists);
      });
    });
  }
});

router.post('/mailing/:projectId', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    var projectId = req.params.projectId;

    var mailingName = req.body.mailing_name;
    var mailingType = req.body.mailing_list_type_radio;
    var mailingEmailAdmin = req.body.mailing_email_admin;
    var mailingPassword = req.body.mailing_password;
    var mailingSubscribePolicy = req.body.subscribe_policy_radio;
    var mailingArchivePolicy = req.body.archive_policy_radio;
    var mailingHost = req.body.mailing_host;

    var newMailingList = {
      "name": mailingName,
      "type": mailingType,
      "admin": mailingEmailAdmin,
      "password": mailingPassword,
      "subscribePolicy": mailingSubscribePolicy,
      "archivePolicy": mailingArchivePolicy,
      "urlhost": mailingHost,
      "emailhost": mailingHost
      // ,"quiet": "TRUE"
    };

    projManagerClient.createMailingList(projectId, newMailingList, function (err, created, mailingListId) {
      if (err) {
        console.log(err);
      }
      return res.redirect('/mailing/' + projectId);
    });
  }
});

router.post('/addParticipantToMailingList/:projectId/', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    var projectId = req.params.projectId;
    var mailinglistName = req.body.mailing_list_name;
    var participant_email = req.body.participant_email;
    var newParticipant = {
      "address": participant_email
    };
    projManagerClient.addParticipantToMailingList(projectId, mailinglistName, newParticipant, function (err, created, response) {
      return res.redirect('/mailing/' + projectId);
    });
  }
});

router.get('/removeParticipantFromMailingList/:projectId/:mailingListName/:participantEmail', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    var projectId = req.params.projectId;
    var mailingListName = req.params.mailingListName;
    var participantTBR = req.params.participantEmail;
    projManagerClient.removeParticipantFromMailingList(projectId, mailingListName, participantTBR, function (err, removed) {
      if (err) {
        console.log("Mailing List [" + mailingListName + "] Error: " + err);
      }
      console.log("Participant [" + participantTBR + "] removed from mailing list [" + mailingListName + "]: " + removed);
      return res.redirect('/mailing/' + projectId);
    });
  }
});

router.get('/removeMailingList/:projectId/:mailingListName', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    var projectId = req.params.projectId;
    var mailingListName = req.params.mailingListName;
    projManagerClient.removeMailingList(projectId, mailingListName, function (err, removed) {
      if (err) {
        console.log("Mailing List [" + mailingListName + "] Error: " + err);
      }
      console.log("Mailing List [" + mailingListName + "] Removed: " + removed);
      return res.redirect('/mailing/' + projectId);
    });
  }
});

module.exports = router;
