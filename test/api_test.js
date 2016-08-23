var api = require("../lib/api");
var assert = require('assert');
var randomstring = require('randomstring');


function randomUserName() {
  return randomstring.generate({
    length: 10,
    charset: 'alphabetic'
  }).toLowerCase();
}

describe('api', function () {

  var apiObj = api("http://localhost:5000");

  describe('Properties', function () {
    describe('apiUrlRoot', function () {
      it('The passed in api root parameter should be available on the returned object', function () {
        assert.equal(apiObj.apiRootUrl, "http://localhost:5000");
      });
    });
  });

  describe('Trusted Auth Endpoints', function () {
    describe('keysForLfId', function () {
      it('Calling keysForLfId with an lfId returns an object with keys', function (done) {
        apiObj.getKeysForLfId("LaneMeyer", function (err, keys) {
          assert.ifError(err);
          assert.equal(keys.keyId.length, 20, "keyId length should be 20");
          assert.equal(keys.secret.length, 40, "secret length should be 40");
          done();
        });
      });
    });
  });

  describe('Admin Endoints', function() {
    describe('POST user/', function(done) {
      var username = randomUserName()
      apiObj.createUser(username, function(err, created) {
        assert.ifError(err);
        assert(created,"New user with username of " + username + " should have been created");
        done();
      });
    });
  });

});