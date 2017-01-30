var api = require("../lib/api");
var assert = require('assert');
var _ = require('underscore');
var randomstring = require('randomstring');

function randomUserName() {
  return randomstring.generate({
    length: 10,
    charset: 'alphabetic'
  }).toLowerCase();
}

describe('api', function () {

  var apiObj = api(process.env['CINCO_SERVER_URL']);

  describe('Properties', function () {
    describe('apiUrlRoot', function () {
      it('The passed in api root parameter should be available on the returned object', function () {
        assert.equal(apiObj.apiRootUrl, process.env['CINCO_SERVER_URL'] + '/');
      });
    });
  });

  describe('Public Endpoints', function () {
    it('about/version', function (done) {
      apiObj.getVersion(function (err, body) {
        assert.equal(body['Application-Name'], 'CINCO');
        done();
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

  describe('Admin Endpoints', function () {
    var adminClient;
    var sampleUserName = randomUserName();

    before(function (done) {
      apiObj.getKeysForLfId("LaneMeyer", function (err, keys) {
        adminClient = apiObj.client(keys);
        adminClient.createUser(sampleUserName, function (err, created) {
          done();
        });
      });
    });

    it('POST users/', function (done) {
      var username = randomUserName();
      adminClient.createUser(username, function (err, created) {
        assert.ifError(err);
        assert(created, "New user with username of " + username + " should have been created");
        done();
      });
    });

    it('GET user/{id}', function (done) {
      adminClient.getUser(sampleUserName, function (err, user) {
        assert.ifError(err);
        assert.equal(user.lfId, sampleUserName, 'Username is not the same as requested');
        assert(user.userId, 'userId property should exist');
        done();
      });
    });

    it('POST user/{id}/group', function (done) {
      var adminGroup = {
        groupId: 2,
        name: 'ADMIN'
      };
      var expected = [{groupId: 1, name: 'USER'}, adminGroup];
      adminClient.addGroupForUser(sampleUserName, adminGroup, function (err, isUpdated, user) {
        assert.ifError(err);
        assert(isUpdated, "User resource should be updated with new group")
        assert.equal(user.lfId, sampleUserName, 'Username is not the same as requested');
        assert.equal(user.groups.length, 2, 'User must have 2 groups');
        assert(_.some(user.groups, function (g) {
          return (g.groupId === adminGroup.groupId) && (g.name === adminGroup.name);
        }));
        assert(_.some(user.groups, function (g) {
          return (g.groupId === 1) && (g.name === 'USER');
        }));
        done();
      });
    });

    it('DELETE users/{id}/group/{groupId}', function (done) {
      var adminGroup = {
        groupId: 2,
        name: 'ADMIN'
      };
      adminClient.addGroupForUser(sampleUserName, adminGroup, function (err, isUpdated, user) {
        assert.ifError(err);
        adminClient.removeGroupFromUser(sampleUserName, adminGroup.groupId, function (err, isUpdated) {
          assert.ifError(err);
          assert(isUpdated);
          adminClient.getUser(sampleUserName, function (err, user) {
            assert.ifError(err);
            assert(!_.some(user.groups, function (g) {
              return g.groupId == adminGroup.groupId;
            }));
            done();
          });
        });
      });
    });

    it('GET usergroups/', function (done) {
      var expected = [{groupId: 1, name: 'USER'}, {groupId: 2, name: 'ADMIN'},
        {groupId: 3, name: 'PROJECT_MANAGER'}];

      adminClient.getAllGroups(function (err, groups) {
        assert.ifError(err);

        _.each(expected, function (eg) {
          var found = _.find(groups, function (g) {
            return (eg.groupId === g.groupId) && (eg.name === g.name);
          });
          assert(found, "Expected group [" + eg + "] not found in returned groups");
        });
        done();
      });
    });

    it('GET users/', function (done) {
      adminClient.getAllUsers(function (err, users, groups) {
        assert.ifError(err);
        assert(users.length >= 2);
        assert(_.some(users, function (u) {
          return u.lfId == 'fvega';
        }));
        assert(_.some(users, function (u) {
          return u.lfId == 'LaneMeyer';
        }));
        done();
      });
    });
  });

  describe('Organizations Endpoints', function () {
    var projManagerClient;
    var projUserName = randomUserName();
    var adminClient;

    before(function (done) {
      apiObj.getKeysForLfId("LaneMeyer", function (err, keys) {
        adminClient = apiObj.client(keys);
        adminClient.createUser(projUserName, function (err, created) {
          var projManagementGroup = {
            groupId: 3,
            name: 'PROJECT_MANAGER'
          };
          adminClient.addGroupForUser(projUserName, projManagementGroup, function (err, updated, user) {
            apiObj.getKeysForLfId(projUserName, function (err, keys) {
              projManagerClient = apiObj.client(keys);
              done();
            });
          });
        });
      });
    });

    it('POST /organizations', function (done) {
      var sampleOrganization = {
        name: "Company Sample Name",
        addresses: [
          {
            type: "MAIN",
            address: {
              country: "US",
              administrativeArea: "Some Province (e.g. Alaska)",
              localityName: "Some City (e.g. Anchorage)",
              postalCode: 99501,
              phone: 800-867-5309,
              thoroughfare: "Some street address"
            }
          },
          {
            type: "BILLING",
            address: {
              country: "US",
              administrativeArea: "Some Province (e.g. Alaska)",
              localityName: "Some City (e.g. Anchorage)",
              postalCode: 99501,
              phone: 800-867-5309,
              thoroughfare: "Some street address"
            }
          }
        ],
        logoRef: "logoName.jpg"
      }
      projManagerClient.createOrganization(sampleOrganization, function (err, created) {
        assert.ifError(err);
        assert(created);
        done();
      });
    });

    it('POST /organizations 403 ', function (done) {
      var username = randomUserName();
      adminClient.createUser(username, function (err) {
        assert.ifError(err);
        apiObj.getKeysForLfId(username, function (err, keys) {
          assert.ifError(err);
          var client = apiObj.client(keys);
          client.createOrganization({}, function (err) {
            assert.equal(err.statusCode, 403);
            done();
          });
        });
      });
    });

    it('POST /organizations 400 on missing name', function (done) {
      var noNameOrg = {
        logoRef: "logoName.jpg"
      };
      projManagerClient.createOrganization(noNameOrg, function (err) {
        assert.equal(err.statusCode, 400);
        done();
      });
    });

    it('GET /organizations', function (done) {
      projManagerClient.getAllOrganizations(function (err, organizations) {
        assert.ifError(err);
        var sampleOrganization = organizations[0];
        assert(sampleOrganization, "A single organization should exist in the returned response array");
        assert(sampleOrganization.id, "id property should exist");
        assert(sampleOrganization.name, "name property should exist");
        assert(sampleOrganization.addresses, "addresses array should exist");
        assert(sampleOrganization.logoRef, "logoRef property should exist");
        done();
      });
    });

    it('GET /organizations/{id}', function (done) {
      projManagerClient.getAllOrganizations(function (err, organizations) {
        assert.ifError(err);
        var id = organizations[0].id;
        projManagerClient.getOrganization(id, function (err, organization) {
          assert.ifError(err);
          assert(organization);
          done();
        });
      });
    });

    it('GET /organizations/{id} 404', function (done) {
      projManagerClient.getAllOrganizations(function (err, organizations) {
        projManagerClient.getOrganization("not_a_real_id", function (err, organization) {
          assert.equal(err.statusCode, 404);
          done();
        });
      });
    });

    it('PUT /organizations/{id}', function (done) {
      projManagerClient.getAllOrganizations(function (err, organizations) {
        assert.ifError(err);
        var organizationId = organizations[0].id;
        var updatedOrganization = {
          id: organizationId,
          name: "Company Updated Name",
          addresses: [
            {
              type: "MAIN",
              address: {
                country: "US",
                administrativeArea: "Some updated Province (e.g. Alaska)",
                localityName: "Some updated City (e.g. Anchorage)",
                postalCode: 99501,
                phone: 888-867-5309,
                thoroughfare: "Some updated street address"
              }
            },
            {
              type: "BILLING",
              address: {
                country: "US",
                administrativeArea: "Some updated Province (e.g. Alaska)",
                localityName: "Some updated City (e.g. Anchorage)",
                postalCode: 99501,
                phone: 888-867-5309,
                thoroughfare: "Some updated street address"
              }
            }
          ],
          logoRef: "logoUpdatedName.jpg"
        }
        projManagerClient.updateOrganization(updatedOrganization, function (err, updated, organization) {
          assert.ifError(err);
          assert(updated);
          assert(organization);
          done();
        });
      });
    });

  });

  describe('Projects Endpoints', function () {
    var projManagerClient;
    var projUserName = randomUserName();
    var adminClient;

    before(function (done) {
      apiObj.getKeysForLfId("LaneMeyer", function (err, keys) {
        adminClient = apiObj.client(keys);
        adminClient.createUser(projUserName, function (err, created) {
          var projManagementGroup = {
            groupId: 3,
            name: 'PROJECT_MANAGER'
          };
          adminClient.addGroupForUser(projUserName, projManagementGroup, function (err, updated, user) {
            apiObj.getKeysForLfId(projUserName, function (err, keys) {
              projManagerClient = apiObj.client(keys);
              done();
            });
          });
        });
      });
    });

    it('GET /projects', function (done) {
      projManagerClient.getAllProjects(function (err, projects) {
        assert.ifError(err);
        var sampleProj = projects[0];
        assert(sampleProj, "A single project should exist in the returned response array");
        assert(sampleProj.id, "id property should exist");
        assert(sampleProj.name, "name property should exist");
        assert(sampleProj.description, "description property should exist");
        assert(sampleProj.pm, "pm property should exist");
        assert(sampleProj.url, "url property should exist");
        assert(sampleProj.startDate, "startDate property should exist");
        assert(_.contains(['DIRECT_FUNDED', 'INCORPORATED', 'UNSPECIFIED'], sampleProj.type),
            "type should be one of: ['DIRECT_FUNDED','INCORPORATED','UNSPECIFIED']. was: " + sampleProj.type);
        done();
      });
    });

    it('GET /projects/{id}', function (done) {
      projManagerClient.getAllProjects(function (err, projects) {
        assert.ifError(err);
        var id = projects[0].id;
        projManagerClient.getProject(id, function (err, project) {
          assert.ifError(err);
          assert(project);
          done();
        });
      })
    });

    it('GET /projects/{id} 404', function (done) {
      projManagerClient.getAllProjects(function (err, projects) {
        projManagerClient.getProject("not_a_real_id", function (err, project) {
          assert.equal(err.statusCode, 404);
          done();
        });
      })
    });

    it('POST /projects', function (done) {
      var sampleProj = {
        name: 'Sample Project',
        description: 'Sample Project Description',
        pm: projUserName,
        url: 'http://www.sample.org/',
        type: 'DIRECT_FUNDED',
        meta: {
          mailingListType : "MM2"
        },
        startDate: '2016-09-26T09:26:36Z'
      };
      projManagerClient.createProject(sampleProj, function (err, created) {
        assert.ifError(err);
        assert(created);
        done();
      });
    });

    it('POST /projects 403 ', function (done) {
      var username = randomUserName();
      adminClient.createUser(username, function (err) {
        assert.ifError(err);
        apiObj.getKeysForLfId(username, function (err, keys) {
          assert.ifError(err);
          var client = apiObj.client(keys);
          client.createProject({}, function (err) {
            assert.equal(err.statusCode, 403);
            done();
          });
        });
      });
    });

    it("POST then GET project", function (done) {
      var sampleProj = {
        name: 'Sample Project',
        description: 'Sample Project Description',
        pm: projUserName,
        url: 'http://www.sample.org/',
        type: 'DIRECT_FUNDED',
        startDate: new Date().toISOString()
      };
      projManagerClient.createProject(sampleProj, function (err, created, id) {
        assert.ifError(err);
        projManagerClient.getProject(id, function (err, returnedProject) {
          assert.ifError(err);
          assert(returnedProject, "A single project should exist in the returned response array");
          assert.equal(returnedProject.name, sampleProj.name);
          assert.equal(returnedProject.description, sampleProj.description);
          assert.equal(returnedProject.pm, sampleProj.pm);
          assert.equal(returnedProject.url, sampleProj.url);
          assert.equal(returnedProject.startDate, sampleProj.startDate);
          done();
        });
      });
    });

    it('DELETE /projects/{id}', function (done) {
      projManagerClient.getMyProjects(function (err, projects) {
        assert.ifError(err);
        var proj = _.last(projects);
        projManagerClient.archiveProject(proj.id, function (err) {
          assert.ifError(err);
          projManagerClient.getProject(proj.id, function (err) {
            assert.equal(err.statusCode, 404);
            done();
          });
        });
      });
    });

    it('PATCH /projects/{id}', function (done) {
      projManagerClient.getMyProjects(function (err, projects) {
        assert.ifError(err);
        var project = _.last(projects);
        var updatedProps = {
          id: project.id,
          pm: projUserName,
          name: randomstring.generate({
            length: 20,
            charset: 'alphabetic'
          }),
          description: randomstring.generate({
            length: 200,
            charset: 'alphabetic'
          })
        };
        projManagerClient.updateProject(updatedProps, function (err, updatedProject) {
          assert.ifError(err);

          assert.equal(updatedProject.id, project.id);
          assert.equal(updatedProject.url, project.url);

          assert.equal(updatedProject.name, updatedProps.name);
          assert.equal(updatedProject.description, updatedProps.description);

          done();
        });
      });
    });

    it('GET /project', function (done) {
      projManagerClient.getMyProjects(function (err, projects) {
        assert.ifError(err);
        var sampleProj = projects[0];
        assert(sampleProj, "A single project should exist in the returned response array");
        assert(sampleProj.id, "id property should exist");
        assert(sampleProj.name, "name property should exist");
        assert(sampleProj.description, "description property should exist");
        assert(sampleProj.pm, "pm property should exist");
        assert(sampleProj.url, "url property should exist");
        assert(sampleProj.startDate, "startDate property should exist");
        assert(_.contains(['DIRECT_FUNDED', 'INCORPORATED', 'UNSPECIFIED'], sampleProj.type),
            "type should be one of: ['DIRECT_FUNDED','INCORPORATED','UNSPECIFIED']. was: " + sampleProj.type);
        done();
      });
    });

    describe('Email Aliases Endpoints', function () {
      it('GET /projects/{id}/emailaliases', function (done) {
        projManagerClient.getAllProjects(function (err, projects) {
          assert.ifError(err);
          var id = projects[0].id;
          projManagerClient.getEmailAliases(id, function (err, emailAliases) {
            assert.ifError(err);
            assert(emailAliases);
            done();
          });
        })
      });

      it('GET /projects/{id}/emailaliases 404', function (done) {
        projManagerClient.getAllProjects(function (err, projects) {
          projManagerClient.getEmailAliases("not_a_real_id", function (err, emailAliases) {
            assert.equal(err.statusCode, 404);
            done();
          });
        })
      });

      it('POST /projects/{id}/emailaliases', function (done) {
        var sampleAlias = {
          "address": "ab@cd.org",
          "participants": [
            {
              "address": "foo@bar.com"
            },
            {
              "address": "foo2@bar.com"
            },
            {
              "address": "foo3@bar.com"
            }
          ]
        };
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.createEmailAliases(projectId, sampleAlias, function (err, created, aliasId) {
            assert.ifError(err);
            assert(created);
            done();
          });
        });
      });

      it('POST /projects/{projectId}/emailaliases/{aliasId}/participants', function (done) {
        var newParticipant = {
          "address": "newFoo@bar.com"
        };
        var sampleAlias = {
          "address": "anEmailAlias@myDomain.org",
          "participants": [
            {
              "address": "foo@bar.com"
            }
          ]
        };
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.createEmailAliases(projectId, sampleAlias, function (err, created, aliasId) {
            assert.ifError(err);
            projManagerClient.addParticipantToEmailAlias(projectId, aliasId.aliasId, newParticipant, function (err, created, response) {
              assert(created);
              done();
            });
          });
        });
      });

      it('DELETE /projects/{projectId}/emailaliases/{aliasId}/participants/{participantAddress}', function (done) {
        var participantTBR = "participantToBeRemoved@bar.com";
        var newParticipant = {
          "address": participantTBR
        };
        var sampleAlias = {
          "address": "anotherEmailAlias@myDomain.org",
          "participants": [
            {
              "address": "someFoo@bar.com"
            }
          ]
        };
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.createEmailAliases(projectId, sampleAlias, function (err, created, aliasId) {
            assert.ifError(err);
            projManagerClient.addParticipantToEmailAlias(projectId, aliasId.aliasId, newParticipant, function (err, created, response) {
              assert.ifError(err);
              projManagerClient.removeParticipantFromEmailAlias(projectId, aliasId.aliasId, participantTBR, function (err, removed) {
                assert(removed);
                done();
              });
            });
          });
        });
      });

    });

    describe('Member Companies Endpoints', function () {
      it('GET /projects/{projectId}/members', function (done) {
        projManagerClient.getAllProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.getMemberCompanies(projectId, function (err, memberCompanies) {
            assert.ifError(err);
            assert(memberCompanies);
            done();
          });
        })
      });

      it('GET /projects/{projectId}/members 404', function (done) {
        projManagerClient.getAllProjects(function (err, projects) {
          projManagerClient.getMemberCompanies("not_a_real_id", function (err, memberCompanies) {
            assert.equal(err.statusCode, 404);
            done();
          });
        })
      });

      it('POST /projects/{projectId}/members', function (done) {
        projManagerClient.getAllOrganizations(function (err, organizations) {
          var organizationId = organizations[0].id;
          var sampleMember = {
            orgId: organizationId,
            tier: {
              type: "PLATINUM",
              qualifier: 1
            },
            startDate: "2016-10-24T15:16:52.885Z",
            renewalDate: "2017-10-24T00:00:00.000Z"
          };
          projManagerClient.getMyProjects(function (err, projects) {
            assert.ifError(err);
            var projectId = projects[0].id;
            projManagerClient.addMemberToProject(projectId, sampleMember, function (err, created, memberId) {
              assert.ifError(err);
              assert(created);
              done();
            });
          });
        });
      });

      it('DELETE /projects/{projectId}/members/{memberId}', function (done) {
        projManagerClient.getAllOrganizations(function (err, organizations) {
          var organizationId = organizations[0].id;
          var memberToBeRemoved = {
            orgId: organizationId,
            tier: {
              type: "GOLD",
            },
            startDate: "2016-03-24T15:16:52.885Z",
            renewalDate: "2017-04-24T00:00:00.000Z"
          };
          projManagerClient.getMyProjects(function (err, projects) {
            assert.ifError(err);
            var projectId = projects[0].id;
            projManagerClient.addMemberToProject(projectId, memberToBeRemoved, function (err, created, memberId) {
              projManagerClient.removeMemberFromProject(projectId, memberId, function (err, removed) {
                assert.ifError(err);
                assert(removed);
                done();
              });
            });
          });
        });
      });

      it('GET /projects/{projectId}/members/{memberId}', function (done) {
        projManagerClient.getAllOrganizations(function (err, organizations) {
          var organizationId = organizations[0].id;
          var sampleMember = {
            orgId: organizationId,
            tier: {
              type: "GOLD",
            },
            startDate: "2016-10-24T15:16:52.885Z",
            renewalDate: "2017-10-24T00:00:00.000Z"
          };
          projManagerClient.getMyProjects(function (err, projects) {
            assert.ifError(err);
            var projectId = projects[0].id;
            projManagerClient.addMemberToProject(projectId, sampleMember, function (err, created, memberId) {
              projManagerClient.getMemberFromProject(projectId, memberId, function (err, memberCompany) {
                assert.ifError(err);
                assert(memberCompany);
                done();
              });
            });
          });
        });
      });

      it('PATCH /projects/{projectId}/members/{memberId}', function (done) {
        projManagerClient.getAllOrganizations(function (err, organizations) {
          var organizationId = organizations[0].id;
          var sampleMember = {
            orgId: organizationId,
            tier: {
              type: "PLATINUM"
            },
            startDate: "2016-10-24T15:16:52.885Z",
            renewalDate: "2017-10-24T00:00:00.000Z"
          };
          var updatedProperties = {
            tier: {
              type: "GOLD"
            },
            renewalDate: "2018-10-24T00:00:00.000Z"
          };
          projManagerClient.getMyProjects(function (err, projects) {
            assert.ifError(err);
            var projectId = projects[0].id;
            projManagerClient.addMemberToProject(projectId, sampleMember, function (err, created, memberId) {
              projManagerClient.updateMember(projectId, memberId, updatedProperties, function (err, updated, updatedMember) {
                assert.ifError(err);
                assert(updatedMember);
                done();
              });
            });
          });
        });
      });

      it('POST /projects/{projectId}/members/{memberId}/contacts', function (done) {
        projManagerClient.getAllOrganizations(function (err, organizations) {
          var organizationId = organizations[0].id;
          var sampleMember = {
            orgId: organizationId,
            tier: {
              type: "GOLD"
            },
            startDate: new Date().toISOString(),
            renewalDate: "2017-10-24T00:00:00.000Z"
          };
          var sampleContact = {
            type: "VOTING",
            givenName: "Grace",
            familyName: "Hopper",
            bio: "Grace Rocks!",
            email: "grace@navy.gov",
            phone: "800-867-5309"
          };
          projManagerClient.getMyProjects(function (err, projects) {
            assert.ifError(err);
            var projectId = projects[0].id;
            projManagerClient.addMemberToProject(projectId, sampleMember, function (err, created, memberId) {
              projManagerClient.addContactToMember(projectId, memberId, sampleContact, function (err, created, contactId) {
                assert.ifError(err);
                assert(created);
                done();
              });
            });
          });
        });
      });

      it('DELETE /projects/{projectId}/members/{memberId}/contacts/{contactId}', function (done) {
        projManagerClient.getAllOrganizations(function (err, organizations) {
          var organizationId = organizations[0].id;
          var sampleMember = {
            orgId: organizationId,
            tier: {
              type: "GOLD",
            },
            startDate: new Date().toISOString(),
            renewalDate: "2017-10-24T00:00:00.000Z"
          };
          var contactToBeRemoved = {
            type: "LEGAL",
            givenName: "Ecarg",
            familyName: "Reppoh",
            bio: "Ecarg Rocks!",
            email: "ecarg@yvan.vog",
            phone: "900-123-4567"
          };
          projManagerClient.getMyProjects(function (err, projects) {
            assert.ifError(err);
            var projectId = projects[0].id;
            projManagerClient.addMemberToProject(projectId, sampleMember, function (err, created, memberId) {
              projManagerClient.addContactToMember(projectId, memberId, contactToBeRemoved, function (err, created, contactId) {
                projManagerClient.removeContactFromMember(projectId, memberId, contactId, function (err, removed) {
                  assert.ifError(err);
                  assert(removed);
                  done();
                });
              });
            });
          });
        });
      });

      it('PUT /projects/{projectId}/members/{memberId}/contacts/{contactId}', function (done) {
        projManagerClient.getAllOrganizations(function (err, organizations) {
          var organizationId = organizations[0].id;
          var sampleMember = {
            orgId: organizationId,
            tier: {
              type: "GOLD",
            },
            startDate: new Date().toISOString(),
            renewalDate: "2019-10-24T00:00:00.000Z"
          };
          var contactToBeUpdated = {
            type: "MARKETING",
            givenName: "Mark",
            familyName: "Eting",
            bio: "Mark Eting Rocks!",
            email: "mark@eting.rock",
            phone: "880-123-4567"
          };
          var updatedContact = {
            type: "FINANCE",
            givenName: "Fin",
            familyName: "Ance",
            bio: "Fin Ance Rocks!",
            email: "fin@ance.rock",
            phone: "990-123-4567"
          };
          projManagerClient.getMyProjects(function (err, projects) {
            assert.ifError(err);
            var projectId = projects[0].id;
            projManagerClient.addMemberToProject(projectId, sampleMember, function (err, created, memberId) {
              projManagerClient.addContactToMember(projectId, memberId, contactToBeUpdated, function (err, created, contactId) {
                projManagerClient.updateContactFromMember(projectId, memberId, contactId, updatedContact, function (err, udpated, contact) {
                  assert.ifError(err);
                  assert(udpated);
                  assert(contact);
                  done();
                });
              });
            });
          });
        });
      });

    });

    describe('Maling Lists Endpoints', function () {

      var mailingListName = randomUserName();

      it('POST /projects/{id}/mailinglists', function (done) {
        var sampleMailingList = {
          "name": mailingListName,
          "type" : "MM2",
          "admin": "admin@domain.org",
          "password": "test_secret_password",
          "subscribePolicy": "OPEN",
          "archivePolicy": "PRIVATE",
          "urlhost": "lists.domain.org",
          "emailhost": "lists.domain.org",
          "quiet": "TRUE"
        };
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.createMailingList(projectId, sampleMailingList, function (err, created, mailingListId) {
            assert.ifError(err);
            assert(created);
            done();
          });
        });
      });

      it('GET /projects/{id}/mailinglists', function (done) {
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.getMailingLists(projectId, function (err, mailingLists) {
            assert.ifError(err);
            assert(mailingLists);
            done();
          });
        })
      });

      it('GET /projects/{id}/mailinglists 404', function (done) {
        projManagerClient.getMailingLists("not_a_real_id", function (err, mailingLists) {
          assert.equal(err.statusCode, 404);
          done();
        });
      });

      var mailingListNameToBeRemoved = randomUserName();

      it('DELETE /projects/{projectId}/mailinglists/{mailinglistId}', function (done) {
        var mailingListToBeRemoved = {
          "name": mailingListNameToBeRemoved,
          "type" : "MM2",
          "admin": "admin@domain.org",
          "password": "test_secret_password",
          "subscribePolicy": "CONFIRM",
          "archivePolicy": "PRIVATE",
          "urlhost": "lists.domain.org",
          "emailhost": "lists.domain.org"
        };
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.createMailingList(projectId, mailingListToBeRemoved, function (err, created, mailinglistId) {
            projManagerClient.removeMailingList(projectId, mailingListNameToBeRemoved, function (err, removed) {
              assert.ifError(err);
              assert(removed);
              done();
            });
          });
        });
      });

      var sampleMailingListName = randomUserName();

      it('GET /projects/{projectId}/mailinglists/{mailinglistId}', function (done) {
        var sampleMailingList = {
          "name": sampleMailingListName,
          "type" : "MM2",
          "admin": "admin@domain.org",
          "password": "test_secret_password",
          "subscribePolicy": "CONFIRM",
          "archivePolicy": "PRIVATE",
          "urlhost": "lists.domain.org",
          "emailhost": "lists.domain.org"
        };
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.createMailingList(projectId, sampleMailingList, function (err, created, mailinglistId) {
            projManagerClient.getMailingListFromProject(projectId, sampleMailingListName, function (err, mailingList) {
              assert.ifError(err);
              assert(mailingList);
              done();
            });
          });
        });
      });

      var sampleParticipantsMailingListName = randomUserName();

      it('POST /projects/{projectId}/mailinglists/{mailinglistId}/participants', function (done) {
        var sampleMailingList = {
          "name": sampleParticipantsMailingListName,
          "type" : "MM2",
          "admin": "admin@domain.org",
          "password": "test_secret_password",
          "subscribePolicy": "OPEN",
          "archivePolicy": "PRIVATE",
          "urlhost": "lists.domain.org",
          "emailhost": "lists.domain.org"
        };
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.createMailingList(projectId, sampleMailingList, function (err, created, mailinglistId) {
            var newParticipant = {
              "address": "participant1@test.com"
            };
            projManagerClient.addParticipantToMailingList(projectId, sampleParticipantsMailingListName, newParticipant, function (err, created, participantEmail) {
              assert.ifError(err);
              assert(created);
              assert(participantEmail);
              done();
            });
          });
        });
      });

      var sampleParticipantsMailingListNameTBR = randomUserName();

      it('DELETE /projects/{projectId}/mailinglists/{mailinglistId}/participants', function (done) {
        var sampleMailingList = {
          "name": sampleParticipantsMailingListNameTBR,
          "type" : "MM2",
          "admin": "admin@domain.org",
          "password": "test_secret_password",
          "subscribePolicy": "OPEN",
          "archivePolicy": "PRIVATE",
          "urlhost": "lists.domain.org",
          "emailhost": "lists.domain.org"
        };
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.createMailingList(projectId, sampleMailingList, function (err, created, mailinglistId) {
            var participantToBeRemoved = {
              "address": "participantTBR@test.com"
            };
            projManagerClient.addParticipantToMailingList(projectId, sampleParticipantsMailingListNameTBR, participantToBeRemoved, function (err, created, participantEmail) {
              projManagerClient.removeParticipantFromMailingList(projectId, sampleParticipantsMailingListNameTBR, participantEmail, function (err, removed) {
                assert.ifError(err);
                assert(removed);
                done();
              });
            });
          });
        });
      });

      var sampleParticipantsMailingListName2 = randomUserName();

      it('GET /projects/{projectId}/mailinglists/{mailinglistId}/participants', function (done) {
        var sampleMailingList = {
          "name": sampleParticipantsMailingListName2,
          "type" : "MM2",
          "admin": "admin@domain.org",
          "password": "test_secret_password",
          "subscribePolicy": "OPEN",
          "archivePolicy": "PRIVATE",
          "urlhost": "lists.domain.org",
          "emailhost": "lists.domain.org"
        };
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.createMailingList(projectId, sampleMailingList, function (err, created, mailinglistId) {
            var newParticipant = {
              "address": "participant1@test.com"
            };
            var newParticipant2 = {
              "address": "participant2@test.com"
            };
            projManagerClient.addParticipantToMailingList(projectId, sampleParticipantsMailingListName2, newParticipant, function (err, created, participantEmail) {
              projManagerClient.addParticipantToMailingList(projectId, sampleParticipantsMailingListName2, newParticipant2, function (err, created, participantEmail) {
                projManagerClient.getParticipantsFromMailingList(projectId, sampleParticipantsMailingListName2, function (err, participants) {
                  assert.ifError(err);
                  assert(participants);
                  done();
                });
              });
            });
          });
        });
      });

      it('GET /projects/{id}/mailinglists && participants', function (done) {
        this.timeout(20000);
        projManagerClient.getMyProjects(function (err, projects) {
          assert.ifError(err);
          var projectId = projects[0].id;
          projManagerClient.getMailingListsAndParticipants(projectId, function (err, mailingLists) {
            assert.ifError(err);
            assert(mailingLists);
            done();
          });
        })
      });

    });
  });
});
