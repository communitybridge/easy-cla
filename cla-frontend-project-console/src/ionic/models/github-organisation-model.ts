export interface GithubOrganisationModel {
    date_created: string;
    date_modified: string;
    organization_company_id: string;
    organization_installation_id: string;
    organization_name: string;
    organization_project_id: string;
    organization_sfid: string;
    version: string;
    providerInfo: {
        avatar_url: string;
        bio: string;
        blog: string;
        company: string;
        created_at: string;
        email: string;
        events_url: string;
        followers: number;
        followers_url: string;
        following: number;
        following_url: string;
        gists_url: string;
        gravatar_id: string;
        hireable: string;
        html_url: string;
        id: number;
        location: string;
        login: string;
        name: string;
        node_id: string;
        organizations_url: string;
        public_gists: number;
        public_repos: number;
        received_events_url: string;
        repos_url: string;
        site_admin: boolean;
        starred_url: string;
        subscriptions_url: string;
        type: string;
        updated_at: string;
        url: string;
    }
    repositories: any[];
}