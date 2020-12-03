FROM 978558897928.dkr.ecr.us-east-1.amazonaws.com/log-extension-image:latest AS layer
FROM public.ecr.aws/lambda/python:3.8
# Layer code
WORKDIR /opt
COPY --from=layer /opt/ .

# Function code
WORKDIR /var/task
COPY app.py .

CMD ["app.lambda_handler"]