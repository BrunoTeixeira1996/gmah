#!/usr/bin/env python3
import sys
import email
from email import policy
from email.parser import BytesParser
from pathlib import Path

def eml_to_html(eml_file):
    with open(eml_file, 'rb') as f:
        msg = BytesParser(policy=policy.default).parse(f)

    html_parts = []
    for part in msg.walk():
        if part.get_content_type() == 'text/html':
            html_parts.append(part.get_content())

    return '\n'.join(html_parts)

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print('Usage: eml_to_html.py <file.eml>')
        sys.exit(1)

    html = eml_to_html(sys.argv[1])
    if html:

        script_dir = Path(__file__).parent

        output_dir = script_dir / 'testdata'
        output_dir.mkdir(exist_ok=True)

        file_name = Path(sys.argv[1]).stem + '.html'
        output_path = output_dir / file_name

        with open(output_path, 'w') as f:
            f.write(html)
            print(f'Generated {file_name} in {output_dir}')
    else:
        print('(No HTML found in this EML file)')
