variable "ecs_cluster_name" {
  description = "The name of the ECS Cluster"
}

variable "external_subnets" {
  description = "External VPC Subnets"
  type = "list"
}

variable "external_elb_sg" {
  description = "Security Group for the external ELB"
}

variable "region" {
  description = "The AWS Region in which we are providing the service in"
  default = "us-west-2"
}

variable "vpc_id" {
  description = "The VPC ID we are deploying to"
}

variable "dns_servers" {
  type = "list"
}

variable "build_hash" {
  description = "The Build Number we are to deploy."
}

variable "ecs_role" {
  description = "The ecsService Role for CINCO"
}

variable "route53_zone_id" {
  description = "The Route53 Zone ID we need to add an entry to for the ALB/ELB."
}

data "template_file" "nginx_ecs_task" {
  template = "${file("${path.module}/nginx-ecs-task.json")}"

  vars {
    # Information for Consul-Template
    CONSUL_SERVICE            = "CLA_CONSOLE_${var.build_hash}"

    # NGINX Domains
    APP_DOMAINS               = "cla.linuxfoundation.org www.cla.linuxfoundation.org"

    # Build Information
    build_hash                = "${var.build_hash}"

    # DNS Servers for Container Resolution
    DNS_SERVER_1              = "${var.dns_servers[0]}"
    DNS_SERVER_2              = "${var.dns_servers[1]}"
    DNS_SERVER_3              = "${var.dns_servers[2]}"

    # Tags for Registrator
    TAG_REGION                = "${var.region}"
    TAG_VPC_ID                = "${var.vpc_id}"
  }
}

resource "aws_ecs_task_definition" "nginx" {
  provider = "aws.local"
  family = "cla-console-nginx"

  lifecycle {
    ignore_changes        = ["image"]
    create_before_destroy = true
  }

  container_definitions = "${data.template_file.nginx_ecs_task.rendered}"
}

resource "aws_ecs_service" "nginx" {
  provider                           = "aws.local"
  name                               = "cla-console-nginx"
  cluster                            = "${var.ecs_cluster_name}"
  task_definition                    = "${aws_ecs_task_definition.nginx.arn}"
  desired_count                      = "3"
  deployment_minimum_healthy_percent = "100"
  deployment_maximum_percent         = "200"
  iam_role                           = "${var.ecs_role}"

  load_balancer {
    target_group_arn   = "${aws_alb_target_group.nginx.arn}"
    container_name     = "nginx"
    container_port     = 80
  }
}

resource "aws_alb_target_group" "nginx" {
  provider             = "aws.local"
  name                 = "cla-console-nginx-80"
  port                 = 80
  protocol             = "HTTP"
  vpc_id               = "${var.vpc_id}"
  deregistration_delay = 30

  health_check {
    path = "/elb-status"
    protocol = "HTTP"
    interval = 15
    matcher = "200,201,202"
  }
}

resource "aws_alb" "nginx" {
  provider           = "aws.local"
  name               = "cla-console-nginx"
  subnets            = ["${var.external_subnets}"]
  security_groups    = ["${var.external_elb_sg}"]
  internal           = false
}

resource "aws_alb_listener" "nginx_80" {
  provider           = "aws.local"
  load_balancer_arn  = "${aws_alb.nginx.id}"
  port               = "80"
  protocol           = "HTTP"

  default_action {
    target_group_arn = "${aws_alb_target_group.nginx.id}"
    type             = "forward"
  }
}

resource "aws_alb_listener" "nginx_443" {
  provider           = "aws.local"
  load_balancer_arn  = "${aws_alb.nginx.id}"
  port               = "443"
  protocol           = "HTTPS"
  certificate_arn    = "arn:aws:acm:us-west-2:643009352547:certificate/cd8b8dbd-dfc4-4417-bee9-b27868d47234"

  default_action {
    target_group_arn = "${aws_alb_target_group.nginx.id}"
    type             = "forward"
  }
}

resource "aws_route53_record" "public" {
  provider= "aws.local"
  zone_id = "${var.route53_zone_id}"
  name    = "."
  type    = "A"

  alias {
    name                   = "${aws_alb.nginx.dns_name}"
    zone_id                = "${aws_alb.nginx.zone_id}"
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "public_www" {
  provider= "aws.local"
  zone_id = "${var.route53_zone_id}"
  name    = "www"
  type    = "A"

  alias {
    name                   = "${aws_alb.nginx.dns_name}"
    zone_id                = "${aws_alb.nginx.zone_id}"
    evaluate_target_health = true
  }
}
