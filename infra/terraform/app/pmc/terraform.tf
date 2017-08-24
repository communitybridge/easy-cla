variable "ecs_cluster_name" {
  description = "The name of the ECS Cluster"
}

variable "internal_subnets" {
  description = "Internal VPC Subnets"
  type = "list"
}

variable "internal_elb_sg" {
  description = "Security Group for the internal ELB"
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

data "template_file" "pmc_ecs_task" {
  template = "${file("${path.module}/pmc-ecs-task.json")}"

  vars {
    # Build Information
    build_hash              = "${var.build_hash}"

    # NGINX Domains
    APP_DOMAINS               = "projectconsole.linuxfoundation.org www.projectconsole.linuxfoundation.org"

    # DNS Servers for Container Resolution
    DNS_SERVER_1              = "${var.dns_servers[0]}"
    DNS_SERVER_2              = "${var.dns_servers[1]}"
    DNS_SERVER_3              = "${var.dns_servers[2]}"

    # Tags for Registrator
    TAG_REGION                = "${var.region}"
    TAG_VPC_ID                = "${var.vpc_id}"
  }
}

resource "aws_ecs_task_definition" "pmc" {
  provider                = "aws.local"
  family                  = "pmc"

  lifecycle {
    ignore_changes        = ["image"]
    create_before_destroy = true
  }

  container_definitions   = "${data.template_file.pmc_ecs_task.rendered}"
}

resource "aws_ecs_service" "pmc" {
  provider                           = "aws.local"
  name                               = "pmc"
  cluster                            = "${var.ecs_cluster_name}"
  task_definition                    = "${aws_ecs_task_definition.pmc.arn}"
  desired_count                      = "3"
  deployment_minimum_healthy_percent = "100"
  deployment_maximum_percent         = "200"
  iam_role                           = "${var.ecs_role}"

  load_balancer {
    target_group_arn   = "${aws_alb_target_group.pmc.arn}"
    container_name     = "pmc"
    container_port     = 80
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_alb_target_group" "pmc" {
  provider             = "aws.local"
  name                 = "pmc-80"
  port                 = 80
  protocol             = "HTTP"
  vpc_id               = "${var.vpc_id}"
  deregistration_delay = 30

  health_check {
    path = "/"
    protocol = "HTTP"
    interval = 15
  }
}

resource "aws_alb" "pmc" {
  provider           = "aws.local"
  name               = "pmc"
  subnets            = ["${var.internal_subnets}"]
  security_groups    = ["${var.internal_elb_sg}"]
}

resource "aws_alb_listener" "pmc" {
  provider           = "aws.local"
  load_balancer_arn  = "${aws_alb.pmc.id}"
  port               = "443"
  protocol           = "HTTPS"
  certificate_arn    = "arn:aws:acm:us-west-2:643009352547:certificate/16db8afa-932a-4bc9-8da4-52c76f00952c"

  default_action {
    target_group_arn = "${aws_alb_target_group.pmc.id}"
    type             = "forward"
  }
}

resource "aws_alb_listener" "pmc_80" {
  provider           = "aws.local"
  load_balancer_arn  = "${aws_alb.pmc.id}"
  port               = "80"
  protocol           = "HTTP"

  default_action {
    target_group_arn = "${aws_alb_target_group.pmc.id}"
    type             = "forward"
  }
}

resource "aws_route53_record" "public" {
  provider= "aws.local"
  zone_id = "${var.route53_zone_id}"
  name    = "."
  type    = "A"

  alias {
    name                   = "${aws_alb.pmc.dns_name}"
    zone_id                = "${aws_alb.pmc.zone_id}"
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "public_www" {
  provider= "aws.local"
  zone_id = "${var.route53_zone_id}"
  name    = "www"
  type    = "A"

  alias {
    name                   = "${aws_alb.pmc.dns_name}"
    zone_id                = "${aws_alb.pmc.zone_id}"
    evaluate_target_health = true
  }
}
