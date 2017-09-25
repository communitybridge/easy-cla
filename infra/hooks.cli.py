import gossip
import lf
import os


@gossip.register('local.instance.init.docker-compose', tags=['cla-console'])
def local_init_docker_compose_file(containers, config, dependencies, envs, mode, path):
    if len(dependencies) >= 1:
        platform_instance = dependencies[0]
        docker_config = lf.utils.loadYaml(os.path.join(platform_instance.path, 'docker-compose.yml'))

        envs = docker_config['services']['workspace']['environment']
        kc_endpoint = [x for k, x in envs.items() if k == 'KEYCLOAK_SERVER_URL'][0]
        kc_instance = platform_instance.dependencies.get('keycloak')

        kc_port = kc_endpoint.split(':')[2]
        containers['workspace']['environment']['KEYCLOAK_SERVER_URL'] = 'http://' + kc_instance.containers.bridge_ip + ':' + kc_port
        lf.logger.info('Setting KEYCLOAK_SERVER_URL to ' + containers['workspace']['environment']['KEYCLOAK_SERVER_URL'])

        for key, port in enumerate(docker_config['services']['workspace']['ports']):
            p = port.split(':')
            if p[1] == '5000':
                containers['workspace']['environment']['CINCO_SERVER_URL'] = 'http://' + platform_instance.containers.bridge_ip + ':' + p[0]
                lf.logger.info('Setting CINCO_SERVER_URL to ' + containers['workspace']['environment']['CINCO_SERVER_URL'])


@gossip.register('preprod_instance_task_build', tags=['cla-console'])
def preprod_instance_task_build(containers, instance_config, dependencies, domains, envs):
    if len(dependencies) >= 1:
        platform = dependencies.get('cinco')

        task = platform.artifacts.get('ECSPreprodTask')
        workspace = [x for x in task.containers if x['name'] == 'workspace'][0]
        kc_endpoint = [x['value'] for x in workspace['environment'] if x['name'] == 'KEYCLOAK_SERVER_URL'][0]

        envs.append({
            'name': 'CINCO_SERVER_URL',
            'value': 'https://' + platform.domain
        })
        envs.append({
            'name': 'KEYCLOAK_SERVER_URL',
            'value': kc_endpoint
        })

        lf.logger.info('Setting CINCO_SERVER_URL to ' + 'https://' + domains['primary'])
        lf.logger.info('Setting KEYCLOAK_SERVER_URL to ' + kc_endpoint)
