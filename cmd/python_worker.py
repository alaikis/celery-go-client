"""
Example Celery worker in Python to work with the Go client.

Install dependencies:
    pip install celery redis

Run worker with Redis:
    celery -A python_worker worker --loglevel=info

Run worker with RabbitMQ:
    celery -A python_worker worker --loglevel=info --broker=amqp://guest:guest@localhost:5672//
"""

from celery import Celery

# Configure Celery to use protocol version 1 (compatible with Go client)
app = Celery(
    'tasks',
    broker='redis://localhost:6379/0',
    backend='redis://localhost:6379/0'
)

# Configure Celery to use JSON serialization and protocol v1
app.conf.update(
    task_serializer='json',
    accept_content=['json'],
    result_serializer='json',
    enable_utc=True,
    task_protocol=1,  # Important: Use protocol version 1
)


@app.task(name='tasks.add')
def add(x, y):
    """Add two numbers"""
    result = x + y
    print(f"Adding {x} + {y} = {result}")
    return result


@app.task(name='tasks.multiply')
def multiply(x, y):
    """Multiply two numbers"""
    result = x * y
    print(f"Multiplying {x} * {y} = {result}")
    return result


@app.task(name='tasks.process_data')
def process_data(name=None, age=None, active=None):
    """Process data with keyword arguments"""
    print(f"Processing data: name={name}, age={age}, active={active}")
    return {
        'name': name,
        'age': age,
        'active': active,
        'processed': True
    }


@app.task(name='tasks.complex_task')
def complex_task(*args, **kwargs):
    """Handle both positional and keyword arguments"""
    print(f"Complex task - Args: {args}, Kwargs: {kwargs}")
    return {
        'args': args,
        'kwargs': kwargs
    }


@app.task(name='tasks.scheduled_task')
def scheduled_task(message):
    """Task that can be scheduled with ETA"""
    print(f"Scheduled task executed: {message}")
    return f"Processed: {message}"


@app.task(name='tasks.priority_task')
def priority_task(message, priority=None):
    """Task for priority queue"""
    print(f"Priority task executed: {message} (priority: {priority})")
    return f"Processed with priority {priority}: {message}"


@app.task(name='tasks.send_email')
def send_email(to=None, subject=None, body=None):
    """Simulate sending an email"""
    print(f"Sending email to {to}")
    print(f"Subject: {subject}")
    print(f"Body: {body}")
    return f"Email sent to {to}"


@app.task(name='tasks.temporary_task')
def temporary_task(data):
    """Task that can expire"""
    print(f"Temporary task executed: {data}")
    return f"Processed: {data}"


@app.task(name='tasks.custom_task')
def custom_task(message):
    """Task for custom queue"""
    print(f"Custom task executed: {message}")
    return f"Processed: {message}"


if __name__ == '__main__':
    # Run worker
    app.worker_main([
        'worker',
        '--loglevel=info',
    ])
