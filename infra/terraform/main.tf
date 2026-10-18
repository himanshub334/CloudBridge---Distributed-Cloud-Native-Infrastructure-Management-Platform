resource "aws_s3_bucket" "state" {
  bucket = var.state_bucket_name
  lifecycle { prevent_destroy = true }
  tags = { Project = var.project_name, Purpose = "terraform-state" }
}
resource "aws_s3_bucket_versioning" "state" {
  bucket = aws_s3_bucket.state.id
  versioning_configuration { status = "Enabled" }
}
resource "aws_s3_bucket_server_side_encryption_configuration" "state" {
  bucket = aws_s3_bucket.state.id
  rule { apply_server_side_encryption_by_default { sse_algorithm = "AES256" } }
}
resource "aws_dynamodb_table" "lock" {
  name = "${var.project_name}-tf-lock"
  billing_mode = "PAY_PER_REQUEST"
  hash_key = "LockID"
  attribute { name = "LockID"; type = "S" }
}
resource "aws_sqs_queue" "dlq" { name = "${var.project_name}-provisioning-dlq"; message_retention_seconds = 1209600 }
resource "aws_sqs_queue" "provisioning" {
  name = "${var.project_name}-provisioning"
  visibility_timeout_seconds = 120
  redrive_policy = jsonencode({ deadLetterTargetArn = aws_sqs_queue.dlq.arn, maxReceiveCount = 4 })
}
resource "aws_cloudwatch_metric_alarm" "queue_depth" {
  alarm_name = "${var.project_name}-queue-depth"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods = 2
  metric_name = "ApproximateNumberOfMessagesVisible"
  namespace = "AWS/SQS"
  period = 300
  statistic = "Average"
  threshold = 100
  dimensions = { QueueName = aws_sqs_queue.provisioning.name }
}
