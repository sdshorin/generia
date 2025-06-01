import logging
import sys
import json
from datetime import datetime
from pythonjsonlogger import jsonlogger

from ..config import LOG_LEVEL

# Logging format configuration
class CustomJsonFormatter(jsonlogger.JsonFormatter):
    def add_fields(self, log_record, record, message_dict):
        super(CustomJsonFormatter, self).add_fields(log_record, record, message_dict)
        
        # Safe timestamp generation - detect if we're in workflow context
        try:
            # First check if we're in a workflow execution context
            from temporalio import workflow
            if workflow.unsafe.is_replaying():
                # We're in a workflow context, use workflow time
                log_record['timestamp'] = workflow.now().isoformat()
            else:
                # We're in an activity or regular context, use system time
                log_record['timestamp'] = datetime.utcnow().isoformat()
        except Exception:
            try:
                # Fallback to workflow time if available
                from temporalio import workflow
                log_record['timestamp'] = workflow.now().isoformat()
            except Exception:
                # Last resort - use a placeholder timestamp
                log_record['timestamp'] = "WORKFLOW_TIME"
        
        log_record['level'] = record.levelname
        log_record['service'] = 'ai-worker'
        
        # Add trace_id and span_id if available
        if hasattr(record, 'trace_id'):
            log_record['trace_id'] = record.trace_id
        if hasattr(record, 'span_id'):
            log_record['span_id'] = record.span_id
    
    def format(self, record):
        # Get the formatted message
        message = super(CustomJsonFormatter, self).format(record)
        
        # Parse the JSON string
        log_data = json.loads(message)
        
        # Process the message field to handle newlines and ensure proper encoding
        if 'message' in log_data:
            # If message is a string, ensure it's properly encoded
            if isinstance(log_data['message'], str):
                # Replace literal \n with actual newlines
                log_data['message'] = log_data['message'].replace('\\n', '\n')
            
            # If message is a dict or other object, ensure it's properly formatted
            if isinstance(log_data['message'], dict):
                # Convert to a properly formatted JSON string with indentation
                log_data['message'] = json.dumps(log_data['message'], ensure_ascii=False, indent=2)
        
        # Convert back to JSON with ensure_ascii=False to preserve non-ASCII characters
        return json.dumps(log_data, ensure_ascii=False)

# Logger configuration
def setup_logger():
    logger = logging.getLogger('ai_worker')
    
    # Set logging level
    level = getattr(logging, LOG_LEVEL.upper(), logging.INFO)
    logger.setLevel(level)
    
    # Add handler for console output
    handler = logging.StreamHandler(sys.stdout)
    formatter = CustomJsonFormatter('%(timestamp)s %(level)s %(service)s %(message)s')
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    
    return logger

# Create global logger instance
logger = setup_logger()

def get_logger():
    """Returns the global logger instance"""
    return logger