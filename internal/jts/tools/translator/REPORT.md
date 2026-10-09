# Translation Report

Stubbed members: 34

| Count | Reason |
| ----- | ------ |
| 10 | runtime junit.TestSuite_new_Class |
| 4 | runtime java.ObjectOutputStream_new_OutputStream |
| 3 | runtime java.Class.GetResourceAsStream_String |
| 3 | runtime java.ObjectInputStream_new_InputStream |
| 2 | constructor reference |
| 2 | runtime java.Class.GetClassLoader |
| 2 | runtime java.IntFunctionFunc |
| 2 | runtime java.Method.Invoke_Object_ObjectArr |
| 1 | runtime java.BinaryOperatorFunc |
| 1 | runtime java.Class.GetConstructor_ClassArr |
| 1 | runtime java.Class.GetDeclaredMethods |
| 1 | runtime java.Class.GetMethod_String_ClassArr |
| 1 | runtime java.Class.IsAssignableFrom_Class |
| 1 | runtime java.Collection_default_ParallelStream |

## Details

### runtime junit.TestSuite_new_Class

- org.locationtech.jts.algorithm.ConvexHullTest.suite
- org.locationtech.jts.geom.GeometryCollectionImplTest.suite
- org.locationtech.jts.geom.GeometryImplTest.suite
- org.locationtech.jts.geom.IntersectionMatrixTest.suite
- org.locationtech.jts.geom.LineStringImplTest.suite
- org.locationtech.jts.geom.MultiPointImplTest.suite
- org.locationtech.jts.geom.PointImplTest.suite
- org.locationtech.jts.geom.PrecisionModelTest.suite
- org.locationtech.jts.io.WKTReaderTest.suite
- org.locationtech.jts.io.WKTWriterTest.suite

### runtime java.ObjectOutputStream_new_OutputStream

- org.locationtech.jts.geom.impl.CoordinateSequenceTestBase.serialize_CoordinateSequence
- org.locationtech.jts.io.SerializabilityTest.testSerializable
- test.jts.geom.TestSerializable.writeData_List
- test.jts.util.SerializationUtil.serialize_Object

### runtime java.Class.GetResourceAsStream_String

- org.locationtech.jts.algorithm.InteriorPointTest.checkInteriorPointResource_String
- org.locationtech.jts.operation.union.CascadedPolygonUnionFileTest.runTestResource_String_double
- test.jts.perf.operation.buffer.FileBufferResultValidatorTest.runTest_String

### runtime java.ObjectInputStream_new_InputStream

- org.locationtech.jts.geom.impl.CoordinateSequenceTestBase.deserialize_byteArr
- test.jts.geom.TestSerializable.readData_List
- test.jts.util.SerializationUtil.deserialize_byteArr

### constructor reference

- test.jts.fail.overlayng.Issue784.<init>
- test.jts.perf.index.FlatbushPerfTest.setUp

### runtime java.Class.GetClassLoader

- org.locationtech.jts.util.TestBuilderProxy.init
- test.jts.TestFiles.getResourceFilePath_String

### runtime java.IntFunctionFunc

- test.jts.perf.geom.prep.PreparedGeometryConcurrencyBugTest.test0
- test.jts.perf.geom.prep.PreparedGeometryConcurrencyBugTest.getFor_double_double

### runtime java.Method.Invoke_Object_ObjectArr

- org.locationtech.jts.util.TestBuilderProxy.showIndicator_Geometry
- org.locationtech.jts.util.TestBuilderProxy.showIndicator_Geometry_Color

### runtime java.BinaryOperatorFunc

- test.jts.fail.overlayng.Issue784.xtestUnionOriginal

### runtime java.Class.GetConstructor_ClassArr

- test.jts.perf.PerformanceTestRunner.runInternal_Class

### runtime java.Class.GetDeclaredMethods

- test.jts.perf.PerformanceTestRunner.findMethods_Class_String

### runtime java.Class.GetMethod_String_ClassArr

- org.locationtech.jts.util.Debug.instancePrintObject_Object

### runtime java.Class.IsAssignableFrom_Class

- org.locationtech.jts.geom.util.GeometryExtracter.toGeometryType_Class

### runtime java.Collection_default_ParallelStream

- test.jts.perf.algorithm.IndexedPointInAreaPerfTest.runParallel
