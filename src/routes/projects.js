if (process.env['NEWRELIC_LICENSE']) require('newrelic');
var express = require('express');
var passport = require('passport');
var request = require('request');
var multer  = require('multer');
var async = require('async');

var cinco = require("../lib/api");

var router = express.Router();

var storage = multer.diskStorage({
  destination: function (req, file, cb) {
    cb(null, 'public/uploads')
  },
  filename: function (req, file, cb) {
    cb(null, file.originalname)
  }
});
var upload = multer({ storage: storage });
var cpUpload = upload.fields([{ name: 'logo', maxCount: 1 }, { name: 'agreement', maxCount: 1 }]);

/**
* Projects:
* Resources to expose and manipulate details of projects
**/

/**
* GET /project
* Look up all projects associated with the loggedIn user
**/
router.get('/project', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
 if(req.session.user.isAdmin || req.session.user.isProjectManager) {
   var projManagerClient = cinco.client(req.session.user.cinco_keys);
   projManagerClient.getMyProjects(function (err, myProjects) {
     req.session.myProjects = myProjects;
     return res.json(myProjects);
   });
 }
});

/**
* GET /project/status
* Get a map of all valid project status values
**/
router.get('/project/status', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
 if(req.session.user.isAdmin || req.session.user.isProjectManager) {
   var projManagerClient = cinco.client(req.session.user.cinco_keys);
   projManagerClient.getProjectStatuses(function (err, statuses) {
     res.send(statuses);
   });
 }
});

/**
* GET /project/categories
* Get a map of all valid project category values
**/
router.get('/project/categories', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
 if(req.session.user.isAdmin || req.session.user.isProjectManager){
   var projManagerClient = cinco.client(req.session.user.cinco_keys);
   projManagerClient.getProjectCategories(function (err, categories) {
     res.send(categories);
   });
 }
});

/**
* GET /project/sectors
* Get a map of all valid project sector values
**/
router.get('/project/sectors', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
 if(req.session.user.isAdmin || req.session.user.isProjectManager){
   var projManagerClient = cinco.client(req.session.user.cinco_keys);
   projManagerClient.getProjectSectors(function (err, sectors) {
     res.send(sectors);
   });
 }
});

/**
* GET /projects
* Look up all projects available
**/
router.get('/projects', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    projManagerClient.getAllProjects(function (err, projects) {
      res.send(projects);
    });
  }
});

/**
* POST /projects
* Add a new project
**/
router.post('/projects', require('connect-ensure-login').ensureLoggedIn('/login'), cpUpload, function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager){

    var now = new Date().toISOString();
    var url = req.body.project_url;
    if(url){
      if (!/^(?:f|ht)tps?\:\/\//.test(url)) url = "http://" + url;
    }
    // var logoFileName = "";
    // var agreementFileName = "";
    // if(req.files){
    //   if(req.files.logo) logoFileName = req.file .logo[0].originalname;
    //   if(req.files.agreement) agreementFileName = req.files.agreement[0].originalname;
    // }

    var newProject = {
      name: req.body.project_name,
      description: req.body.project_description,
      managers: [req.session.user.user],
      url: url,
      sector: req.body.project_sector,
      address: JSON.parse(req.body.project_address),
      status: req.body.project_status,
      category: req.body.project_category,
      startDate: req.body.project_start_date?req.body.project_start_date:now
    };

    console.log(newProject);

    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    projManagerClient.createProject(newProject, function (err, created, projectId) {
      return res.json(projectId);
    });
  }
});

/**
* DELETE /projects/{projectId}
* Archive a project
**/
// router.delete('/projects/:projectId', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
//   if(req.session.user.isAdmin || req.session.user.isProjectManager){
//     var projectId = req.params.projectId;
//     var projManagerClient = cinco.client(req.session.user.cinco_keys);
//     projManagerClient.archiveProject(projectId, function (err) {
//       return res.redirect('/');
//     });
//   }
// });

/**
* GET /projects/{projectId}
* Get an individual project
**/
router.get('/projects/:projectId', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
 if(req.session.user.isAdmin || req.session.user.isProjectManager){
   var projectId = req.params.projectId;
   if(req.query.members) { projectId = projectId + '?members=' + req.query.members }
   var projManagerClient = cinco.client(req.session.user.cinco_keys);
   projManagerClient.getProject(projectId, function (err, project) {
     // TODO: Create 404 page for when project doesn't exist
     if (err) return res.send('');
     res.send(project);
   });
 }
});

/**
* TODO: Convert to PUT /projects/{projectId}
* POST /edit_project/{projectId}
* Update a project by id
**/
router.post('/edit_project/:projectId', require('connect-ensure-login').ensureLoggedIn('/login'), cpUpload, function(req, res){
 if(req.session.user.isAdmin || req.session.user.isProjectManager){

   var projectId = req.params.projectId;

   // var logoFileName = "";
   // var agreementFileName = "";
   // var url = req.body.url;
   // if(url){
     // if (!/^(?:f|ht)tps?\:\/\//.test(url)) url = "http://" + url;
   // }
   // if(req.files.logo) logoFileName = req.files.logo[0].originalname;
   // else logoFileName = req.body.old_logoRef;
   // if(req.files.agreement) agreementFileName = req.files.agreement[0].originalname;
   // else agreementFileName = req.body.old_agreementRef;

   var updatedProps = {
     id: projectId,
     name: req.body.project_name,
     description: req.body.project_description,
     url: req.body.project_url,
     sector: req.body.project_sector,
     address: JSON.parse(req.body.project_address),
     status: req.body.project_status,
     category: req.body.project_category,
     startDate: req.body.project_start_date
     // pm: req.body.creator_pm,
     // logoRef: logoFileName,
     // agreementRef: agreementFileName,
   };
   var projManagerClient = cinco.client(req.session.user.cinco_keys);
   projManagerClient.updateProject(updatedProps, function (err, updatedProject) {
     return res.json(updatedProject);
   });
 }
});


/**
* GET /projects/{projectId}/config
* Retrieve the configuration associated with this project
**/
router.get('/projects/:projectId/config', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager) {
    var projectId = req.params.projectId;
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    projManagerClient.getProjectConfig(projectId, function (err, config) {
      // TODO: Create 404 page for when project doesn't exist
      if (err) return res.send('');
      res.send(config);
    });
  }
});

/**
* PUT /projects/{projectId}/managers
* Update the user ids that should be program managers of this project
**/
router.put('/projects/:projectId/managers', require('connect-ensure-login').ensureLoggedIn('/login'), cpUpload, function(req, res){
  if(req.session.user.isAdmin || req.session.user.isProjectManager) {
    var projectId = req.params.projectId;
    var managers = JSON.parse(req.body.managers);
    var projManagerClient = cinco.client(req.session.user.cinco_keys);
    projManagerClient.updateProjectManagers(projectId, managers, function (err, projectConfig) {
      return res.json(projectConfig);
    });
  }
});

module.exports = router;
