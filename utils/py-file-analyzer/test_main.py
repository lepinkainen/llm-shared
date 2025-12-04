#!/usr/bin/env python3
"""Tests for py-file-analyzer."""

import os
import tempfile
import unittest
from unittest.mock import patch

# Add parent directory to path to import the analyzer
import sys
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from main import analyze_file, collect_python_files, _calculate_function_buckets


class TestPyFileAnalyzer(unittest.TestCase):
    def setUp(self):
        """Create a temporary directory with test Python files."""
        self.test_dir = tempfile.mkdtemp()
        
        # Simple test file
        simple_file = os.path.join(self.test_dir, "simple.py")
        with open(simple_file, 'w') as f:
            f.write("def hello():\n    print('Hello')\n\n")
            f.write("class Test:\n    def method(self):\n        pass\n")
        
        # Complex test file
        complex_file = os.path.join(self.test_dir, "complex.py")
        with open(complex_file, 'w') as f:
            f.write("import os\nimport sys\n\n")
            f.write("async def long_function():\n")
            for i in range(60):
                f.write(f"    print('Line {i}')\n")
            f.write("    return 'done'\n\n")
            f.write("class Complex:\n")
            f.write("    def method1(self):\n        pass\n")
            f.write("    def method2(self):\n        pass\n")
            
        # __init__.py file (should be excluded)
        init_file = os.path.join(self.test_dir, "__init__.py")
        with open(init_file, 'w') as f:
            f.write("# Package initialization file\n")
    
    def tearDown(self):
        """Clean up test files."""
        import shutil
        shutil.rmtree(self.test_dir)
    
    def test_simple_file_analysis(self):
        """Test analysis of a simple Python file."""
        simple_file = os.path.join(self.test_dir, "simple.py")
        analysis = analyze_file("simple.py", self.test_dir)
        
        self.assertEqual(analysis.path, "simple.py")
        self.assertGreater(analysis.lines, 0)
        self.assertEqual(len(analysis.classes), 1)  # Test class
        self.assertEqual(len(analysis.functions), 2)  # hello function and method
        self.assertEqual(analysis.top_level_functions, 1)  # just hello()
        self.assertEqual(analysis.method_counts["Test"], 1)  # one method in Test
    
    def test_complex_file_analysis(self):
        """Test analysis of a more complex Python file."""
        complex_file = os.path.join(self.test_dir, "complex.py")
        analysis = analyze_file("complex.py", self.test_dir)
        
        self.assertEqual(analysis.path, "complex.py")
        self.assertGreater(analysis.lines, 60)
        self.assertEqual(len(analysis.classes), 1)  # Complex class
        self.assertEqual(len(analysis.functions), 3)  # long_function + 2 methods
        self.assertEqual(analysis.top_level_functions, 1)  # long_function
        
        # Check that we detected the long function
        long_func = next(f for f in analysis.functions if f.name == "long_function")
        self.assertTrue(long_func.is_async)
        self.assertGreater(long_func.lines, 50)
    
    def test_calculate_function_buckets(self):
        """Test function bucket calculation."""
        from main import FunctionInfo
        
        functions = [
            FunctionInfo("func1", 10, 5, False, False),
            FunctionInfo("func2", 30, 15, False, False),
            FunctionInfo("func3", 75, 25, False, False),
            FunctionInfo("func4", 120, 40, False, False),
        ]
        
        buckets = _calculate_function_buckets(functions)
        
        self.assertEqual(buckets['under20'], 1)  # func1
        self.assertEqual(buckets['between20_49'], 1)  # func2
        self.assertEqual(buckets['between50_100'], 1)  # func3
        self.assertEqual(buckets['over100'], 1)  # func4
    
    def test_collect_python_files(self):
        """Test collecting Python files from a directory."""
        # Test with exclude_tests=True (default)
        analyses = collect_python_files(self.test_dir, exclude_tests=True)
        
        # Should have 2 Python files (simple.py, complex.py, excluding __init__.py and test files)
        self.assertEqual(len(analyses), 2)
        
        # Check that we have expected files
        paths = [a.path for a in analyses]
        self.assertIn("simple.py", paths)
        self.assertIn("complex.py", paths)
        self.assertNotIn("__init__.py", paths)  # Should be excluded
        
        # Test with exclude_tests=False
        # Create a test file in the test directory
        test_file = os.path.join(self.test_dir, "test_example.py")
        with open(test_file, 'w') as f:
            f.write("import unittest\n\nclass TestExample(unittest.TestCase):\n    def test_something(self):\n        self.assertTrue(True)\n")
        
        # Should still exclude test files when exclude_tests=True
        analyses_no_tests = collect_python_files(self.test_dir, exclude_tests=True)
        self.assertEqual(len(analyses_no_tests), 2)  # Still 2, test file excluded
        
        # Should include test files when exclude_tests=False
        analyses_with_tests = collect_python_files(self.test_dir, exclude_tests=False)
        self.assertEqual(len(analyses_with_tests), 3)  # Now 3, test file included
        # But __init__.py should still be excluded
        paths_with_tests = [a.path for a in analyses_with_tests]
        self.assertNotIn("__init__.py", paths_with_tests)


if __name__ == '__main__':
    unittest.main()