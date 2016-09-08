var express = require('express');
var passport = require('passport');
var dummy_data = require('./dummy_db/dummy_data');
var request = require('request');
var cinco_api = require("./lib/api");

var router = express.Router();

const integration_user = process.env['CONSOLE_INTEGRATION_USER'];
const integration_pass = process.env['CONSOLE_INTEGRATION_PASSWORD'];

var hostURL = process.env['CINCO_SERVER_URL'];
if(process.argv[2] == 'dev') hostURL = 'http://localhost:5000';
if(!hostURL.startsWith('http')) hostURL = 'http://' + hostURL;
console.log("hostURL: " + hostURL);

var cinco = cinco_api(hostURL);

router.get('/', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  res.render('homepage');
});

router.get('/angular', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  res.render('angular');
});

router.get('/logout', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  req.session.user = '';
  req.session.destroy();
  req.logout();
  res.redirect('/');
});

router.get('/login', function(req,res) {
  res.render('login');
});

router.get('/404', function(req,res) {
  res.render('404', { lfid: "" });
});

router.get('/401', function(req,res) {
  res.render('401', { lfid: "" });
});

router.get('/login_cas', function(req, res, next) {
  passport.authenticate('cas', function (err, user, info) {
    if (err) return next(err);
    if(user)
    {
      req.session.user = user;
    }
    if (!user) {
      req.session.destroy();
      return res.redirect('/login');
    }
    req.logIn(user, function (err) {
      if (err) return next(err);
      var lfid = req.session.user.user;
      cinco.getKeysForLfId(lfid, function (err, keys) {
        if(keys){
          req.session.user.cinco_keys = keys;
          var adminClient = cinco.client(keys);
          adminClient.getUser(lfid, function(err, user) {
            if(user){
              req.session.user.cinco_groups = JSON.stringify(user.groups);
              req.session.user.isAdmin = false;
              req.session.user.isUser = false;
              req.session.user.isProjectManager = false;
              if(user.groups)
              {
                if(user.groups.length > 0){
                  for(var i = 0; i < user.groups.length; i ++)
                  {
                    if(user.groups[i].name == "ADMIN") req.session.user.isAdmin = true;
                    if(user.groups[i].name == "USER") req.session.user.isUser = true;
                    if(user.groups[i].name == "PROJECT_MANAGER") req.session.user.isProjectManager = true;
                  }
                }
              }
              return res.redirect('/');
            }
            else {
              return res.redirect('/login');
            }
          });
        }
        if(err){
          req.session.destroy();
          console.log("getKeysForLfId err: " + err);
          if(err.statusCode == 404) return res.render('404', { lfid: lfid }); // Returned if a user with the given id is not found
          if(err.statusCode == 401) return res.render('401', { lfid: lfid }); // Unable to get keys for lfid given. User unauthorized.
        }
      });
    });
  })(req, res, next);
});

router.get('/profile', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  var adminClient = cinco.client(req.session.user.cinco_keys);
  var lfid = req.session.user.user;
  adminClient.getUser(lfid, function(err, user) {
    if(user){
      req.session.user.cinco_groups = JSON.stringify(user.groups);
      res.render('profile');
    }
    else {
      res.render('profile');
    }
  });
});

router.get('/create_project', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  res.render('create_project');
});

router.get('/add_company', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  res.render('add_company');
});

router.get('/project', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  dummy_data.findProjectById(req.query.id, function(err, project_data) {
    if(project_data) res.render('project', { project_data: project_data });
    else res.redirect('/');
  });
});

router.get('/mailing', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  dummy_data.findProjectById(req.query.id, function(err, project_data) {
    if(project_data) res.render('mailing', { project_data: project_data });
    else res.redirect('/');
  });
});

router.get('/aliases', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  dummy_data.findProjectById(req.query.id, function(err, project_data) {
    if(project_data) res.render('aliases', { project_data: project_data });
    else res.redirect('/');
  });
});

router.get('/members', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  dummy_data.findProjectById(req.query.id, function(err, project_data) {
    if(project_data) res.render('members', { project_data: project_data });
    else res.redirect('/');
  });
});

router.get('/admin', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin) {
    var adminClient = cinco.client(req.session.user.cinco_keys);
    var username = req.body.form_lfid;
    adminClient.getAllUsers(function (err, users) {
      res.render('admin', { message: "", users: users });
    });
  }
  else res.redirect('/');
});

router.post('/create_project_manager_user', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin){
    var adminClient = cinco.client(req.session.user.cinco_keys);
    var username = req.body.form_lfid;
    var projectManagerGroup = {
      groupId: 3,
      name: 'PROJECT_MANAGER'
    }
    var userGroup = {
      groupId: 1,
      name: 'USER'
    }
    adminClient.createUser(username, function (err, created) {
      var message = '';
      if (err) {
        message = err;
        adminClient.getAllUsers(function (err, users) {
          return res.render('admin', { message: message, users: users });
        });
      }
      if(created) {
        message = 'Project Manager has been created.';
        adminClient.addGroupForUser(username, userGroup, function(err, isUpdated, user) {});
        adminClient.addGroupForUser(username, projectManagerGroup, function(err, isUpdated, user) {
          if (err) message = err;
          adminClient.getAllUsers(function (err, users) {
            return res.render('admin', { message: message, users: users });
          });
        });
      }
      else {
        message = 'User already exists.';
        adminClient.addGroupForUser(username, userGroup, function(err, isUpdated, user) {});
        adminClient.addGroupForUser(username, projectManagerGroup, function(err, isUpdated, user) {
          message = 'User already exists. ' + 'Project Manager has been created.';
          if (err) message = err;
          adminClient.getAllUsers(function (err, users) {
            return res.render('admin', { message: message, users: users });
          });
        });
      }
    });
  }
});

router.post('/create_admin_user', require('connect-ensure-login').ensureLoggedIn('/login'), function(req, res){
  if(req.session.user.isAdmin){
    var adminClient = cinco.client(req.session.user.cinco_keys);
    var username = req.body.form_lfid;
    var adminGroup = {
      groupId: 2,
      name: 'ADMIN'
    }
    var userGroup = {
      groupId: 1,
      name: 'USER'
    }
    adminClient.createUser(username, function (err, created) {
      var message = '';
      if (err) {
        message = err;
        adminClient.getAllUsers(function (err, users) {
          return res.render('admin', { message: message, users: users });
        });
      }
      if(created) {
        message = 'Admin has been created.';
        adminClient.addGroupForUser(username, userGroup, function(err, isUpdated, user) {});
        adminClient.addGroupForUser(username, adminGroup, function(err, isUpdated, user) {
          if (err) message = err;
          adminClient.getAllUsers(function (err, users) {
            return res.render('admin', { message: message, users: users });
          });
        });
      }
      else {
        message = 'User already exists. ' + 'Admin has been created.';
        adminClient.addGroupForUser(username, userGroup, function(err, isUpdated, user) {});
        adminClient.addGroupForUser(username, adminGroup, function(err, isUpdated, user) {
          if (err) message = err;
          adminClient.getAllUsers(function (err, users) {
            return res.render('admin', { message: message, users: users });
          });
        });
      }
    });
  }
});

router.get('*', function(req, res) {
    res.redirect('/');
});

module.exports = router;
