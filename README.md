AI Agent CLI – Terminal Assistant for Developers
AI Agent CLI là một trợ lý dòng lệnh giúp tự động hóa các thao tác hệ thống thông qua ngôn ngữ tự nhiên. Ứng dụng sử dụng trí tuệ nhân tạo để hiểu và thực thi các lệnh như quản lý thư mục, kết nối SSH, thao tác với Git, ghi nhớ thông tin cá nhân, và nhiều tác vụ phát triển phần mềm thường gặp.

Key Features
Hiểu và phân loại câu lệnh từ ngôn ngữ tự nhiên

Tạo, cập nhật, hoặc xóa thư mục thông qua mô tả ngôn ngữ

Tự động kết nối SSH vào EC2 dựa trên memory đã lưu (key, username, IP)

Ghi nhớ các thông tin người dùng: thư mục, server, đường dẫn, v.v.

Trả lời câu hỏi dựa trên thông tin đã lưu hoặc yêu cầu người dùng bổ sung khi thiếu

Cho phép mở rộng thêm các chức năng như thao tác với Git, mở ứng dụng, kiểm tra hệ thống, v.v.

Planned Features
Xóa file và thư mục qua ngôn ngữ tự nhiên

Ghi nhớ thông tin cá nhân trực tiếp từ đoạn chat

Clone hoặc pull GitHub repository về thư mục cụ thể

Kiểm tra ứng dụng đang chiếm nhiều tài nguyên hệ thống

Lấy số lượng Pull Request đang mở bằng GitHub API

Tạo và ghi nội dung vào file bất kỳ với phần mở rộng tuỳ chọn

Khởi động ứng dụng như Google Chrome qua lệnh thoại hoặc dòng lệnh

Example Commands
bash
sewn@Mac-mini-cua-Nguyen SwxProject % cd ai-agent-go 
sewn@Mac-mini-cua-Nguyen ai-agent-go % go build .    
sewn@Mac-mini-cua-Nguyen ai-agent-go % go run .      
Hi user!
>>> Where is the location of T1 folder ?
Classifying input:  Question
✅ Answer from memory: The location of the T1 folder is at /Users/sewn/Desktop/T1.

Requirements
Go 1.20+

API tương thích OpenAI hoặc mô hình cục bộ như Mistral, DeepSeek, Qwen

Hệ điều hành macOS hoặc Linux (ưu tiên macOS trong phát triển hiện tại)

Setup
bash
git clone https://github.com/yourusername/ai-agent-go.git
cd ai-agent-go
go run .
Model Configuration
Để thay đổi mô hình AI sử dụng, bạn có thể chỉnh hàm AskLLM() trong tệp:
internal/llm/ask.go
Định hướng phát triển
Dự án hướng đến việc tạo ra một trợ lý CLI linh hoạt, có khả năng học và thích nghi với ngữ cảnh làm việc của từng cá nhân, từ thao tác hệ thống đến hỗ trợ quản lý code và hạ tầng. Mục tiêu là giúp developer tiết kiệm thời gian, giảm thao tác lặp đi lặp lại, và tăng cường hiệu suất làm việc qua giao tiếp ngôn ngữ tự nhiên.

