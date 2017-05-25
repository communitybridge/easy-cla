/**
 * Creates basic security groups to be used by instances and ELBs.
 */

variable "name" {
  description = "The name of the security groups serves as a prefix, e.g stack"
}

variable "vpc_id" {
  description = "The VPC ID"
}

variable "cidr" {
  description = "The cidr block to use for internal security groups"
}

# Security Group for EC2 Instances under our ECS Cluster.
resource "aws_security_group" "ecs" {
  provider    = "aws.local"
  name        = "${var.name}-ecs-cluster"
  description = "PMC - ECS Container Instances"
  vpc_id      = "${var.vpc_id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    security_groups = ["${aws_security_group.external_elb.id}", "${aws_security_group.internal_ssh.id}"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name        = "PMC - ECS Container Instances"
  }
}

# Used for the NGINX Load Balancer.
resource "aws_security_group" "external_elb" {
  provider    = "aws.local"
  name        = "${var.name}-external-elb"
  description = "Allows ELB Access Access"
  vpc_id      = "${var.vpc_id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags {
    Name        = "PMC - External ELB"
  }
}

# Used for the PMC Load Balancer (only used for HealthChecks)
resource "aws_security_group" "internal_elb" {
  provider    = "aws.local"
  name        = "${var.name}-internal-elb"
  description = "Allows ELB Internal Access"
  vpc_id      = "${var.vpc_id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["${var.cidr}"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["${var.cidr}"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags {
    Name        = "PMC - Internal ELB"
  }
}

# Allow internal SSH Access from within the Engineering Pool.
resource "aws_security_group" "internal_ssh" {
  provider    = "aws.local"
  name        = "${var.name}-internal-ssh"
  description = "Allows ssh from the VPN"
  vpc_id      = "${var.vpc_id}"

  ingress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    cidr_blocks     = ["10.32.0.0/12"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["${var.cidr}"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags {
    Name        = "PMC - VPN SSH Access"
  }
}

output "internal_ssh" {
  value = "${aws_security_group.internal_ssh.id}"
}

output "external_elb" {
  value = "${aws_security_group.external_elb.id}"
}

output "internal_elb" {
  value = "${aws_security_group.internal_elb.id}"
}

output "ecs-cluster" {
  value = "${aws_security_group.ecs.id}"
}