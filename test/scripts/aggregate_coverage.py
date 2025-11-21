#!/usr/bin/env python3
"""
Coverage aggregation script for multi-language SDK tests.
Parses coverage reports from different languages and generates a combined report.
"""

import os
import sys
import json
import xml.etree.ElementTree as ET
from pathlib import Path
from typing import Dict, Any

COVERAGE_THRESHOLD = 80.0


def parse_python_coverage(results_dir: Path) -> Dict[str, Any]:
    """Parse Python XML coverage report"""
    python_xml = results_dir / 'coverage-python.xml'
    if not python_xml.exists():
        print(f"⚠️  Python coverage file not found: {python_xml}")
        return {'lines': 0, 'covered': 0, 'percentage': 0}

    try:
        tree = ET.parse(python_xml)
        root = tree.getroot()
        lines = int(root.attrib.get('lines-valid', 0))
        covered = int(root.attrib.get('lines-covered', 0))
        percentage = (covered / lines * 100) if lines > 0 else 0

        return {
            'lines': lines,
            'covered': covered,
            'percentage': round(percentage, 2)
        }
    except Exception as e:
        print(f"❌ Error parsing Python coverage: {e}")
        return {'lines': 0, 'covered': 0, 'percentage': 0}


def parse_nodejs_coverage(results_dir: Path) -> Dict[str, Any]:
    """Parse Node.js JSON coverage report"""
    nodejs_json = results_dir / 'coverage-nodejs' / 'coverage-summary.json'
    if not nodejs_json.exists():
        print(f"⚠️  Node.js coverage file not found: {nodejs_json}")
        return {'lines': 0, 'covered': 0, 'percentage': 0}

    try:
        with open(nodejs_json) as f:
            data = json.load(f)
            total = data.get('total', {})
            lines = total.get('lines', {})

            total_lines = lines.get('total', 0)
            covered_lines = lines.get('covered', 0)
            percentage = lines.get('pct', 0)

            return {
                'lines': total_lines,
                'covered': covered_lines,
                'percentage': round(percentage, 2)
            }
    except Exception as e:
        print(f"❌ Error parsing Node.js coverage: {e}")
        return {'lines': 0, 'covered': 0, 'percentage': 0}


def parse_openapi_coverage(results_dir: Path) -> Dict[str, Any]:
    """Check OpenAPI endpoint coverage"""
    # TODO: Implement actual endpoint coverage checking
    return {
        'total_endpoints': 0,
        'tested_endpoints': 0,
        'percentage': 0
    }


def generate_report(results_dir: Path) -> int:
    """Generate combined coverage report"""

    print("\n" + "="*70)
    print("📊 SDK COVERAGE REPORT")
    print("="*70)

    coverage_data = {
        'python': parse_python_coverage(results_dir),
        'nodejs': parse_nodejs_coverage(results_dir),
    }

    # Calculate total coverage
    total_lines = sum(lang['lines'] for lang in coverage_data.values())
    covered_lines = sum(lang['covered'] for lang in coverage_data.values())
    total_percentage = (covered_lines / total_lines * 100) if total_lines > 0 else 0

    # Print language-specific coverage
    for lang, data in coverage_data.items():
        icon = "✅" if data['percentage'] >= COVERAGE_THRESHOLD else "⚠️ "
        print(f"\n{lang.upper()}:")
        print(f"  Lines:    {data['covered']}/{data['lines']}")
        print(f"  Coverage: {data['percentage']:.2f}% {icon}")

    # Print total coverage
    print(f"\n{'─'*70}")
    total_icon = "✅" if total_percentage >= COVERAGE_THRESHOLD else "❌"
    print(f"TOTAL COVERAGE: {total_percentage:.2f}% {total_icon}")
    print(f"Threshold:      {COVERAGE_THRESHOLD}%")
    print("="*70)

    # Write summary JSON
    summary = {
        'languages': coverage_data,
        'total': {
            'lines': total_lines,
            'covered': covered_lines,
            'percentage': round(total_percentage, 2)
        },
        'threshold': COVERAGE_THRESHOLD,
        'passed': total_percentage >= COVERAGE_THRESHOLD
    }

    summary_file = results_dir / 'coverage-summary.json'
    with open(summary_file, 'w') as f:
        json.dump(summary, f, indent=2)
    print(f"\n📄 Coverage summary written to: {summary_file}")

    # Check if tests passed
    if total_percentage < COVERAGE_THRESHOLD:
        print(f"\n❌ FAILURE: Coverage {total_percentage:.2f}% is below threshold {COVERAGE_THRESHOLD}%")
        return 1
    else:
        print(f"\n✅ SUCCESS: Coverage {total_percentage:.2f}% meets threshold {COVERAGE_THRESHOLD}%")
        return 0


def main():
    if len(sys.argv) < 2:
        print("Usage: aggregate_coverage.py <results_dir>")
        sys.exit(1)

    results_dir = Path(sys.argv[1])

    if not results_dir.exists():
        print(f"❌ Results directory not found: {results_dir}")
        sys.exit(1)

    exit_code = generate_report(results_dir)
    sys.exit(exit_code)


if __name__ == '__main__':
    main()
