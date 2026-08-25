output "public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.app.public_ip
}

output "app_url" {
  description = "URL to open the deployed sentinel dashboard"
  value       = "http://${aws_instance.app.public_ip}:8090"
}

output "ssh_command" {
  description = "Command to SSH into the instance"
  value       = "ssh -i ${local_sensitive_file.private_key.filename} ubuntu@${aws_instance.app.public_ip}"
}
